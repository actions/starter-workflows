#!/usr/bin/env npx ts-node
import { promises as fs } from "fs";
import { load } from "js-yaml";
import { extname, join } from "path";

async function findActionsInWorkflows(folders: string[]): Promise<Map<string, string[]>> {
  const jobUses: string[] = [];
  const stepUses: string[] = [];
  
  for (const folder of folders) {
    const dir = await fs.readdir(folder, {
      withFileTypes: true,
    });

    for (const entry of dir) {
      if (entry.isFile() && [".yml", ".yaml"].includes(extname(entry.name))) {
        try {
          const workflowFileContent = await fs.readFile(join(folder, entry.name), "utf8");
          const workflow: any = load(workflowFileContent);
          
          if (workflow.jobs) {
            for (const jobName in workflow.jobs) {
              const job = workflow.jobs[jobName];
              
              if (job.uses) {
                jobUses.push(job.uses);
              }
              
              if (job.steps) {
                for (const step of job.steps) {
                  if (step.uses) {
                    stepUses.push(step.uses);
                  }
                }
              }
            }
          }
        } catch (error) {
          console.error(`Error reading ${join(folder, entry.name)}:`, error);
        }
      }
    }
  }

  return new Map([
    ["JOB_USES", jobUses],
    ["JOBS_STEP_USES", stepUses]
  ]);
}

const GITHUB_TOKEN = process.env.GITHUB_TOKEN || process.env.GH_TOKEN || "";

function getHeaders(): Record<string, string> {
  const headers: Record<string, string> = {
    "Accept": "application/vnd.github.v3+json"
  };
  if (GITHUB_TOKEN) {
    headers["Authorization"] = `Bearer ${GITHUB_TOKEN}`;
  }
  return headers;
}

async function getCommitForRef(owner: string, repo: string, ref: string): Promise<string | null> {
  try {
    const apiUrl = `https://api.github.com/repos/${owner}/${repo}/commits/${ref}`;
    const response = await fetch(apiUrl, { headers: getHeaders() });
    
    if (response.status === 403) {
      console.error(`\nGitHub API rate limit exceeded!`);
      console.error(`Unauthenticated: 60 requests/hour | Authenticated: 5,000 requests/hour`);
      console.error(`\nCreate a token at: https://github.com/settings/tokens/new?description=check-outdated-actions&scopes=public_repo`);
      console.error(`Then set: export GITHUB_TOKEN="your_token"`);
      process.exit(1);
    }
    
    if (response.status !== 200) {
      return null;
    }
    
    const data = await response.json();
    return data.sha || null;
  } catch {
    return null;
  }
}

async function getLatestRelease(actionRef: string): Promise<{ tag: string; commit: string } | null> {
  try {
    // Parse the action reference (e.g., "actions/checkout@v5" or "owner/repo/.github/workflows/workflow.yml@v1")
    const [fullAction] = actionRef.split("@");
    const parts = fullAction.split("/");
    
    if (parts.length < 2) {
      return null;
    }
    
    const owner = parts[0];
    const repo = parts[1];
    
    // Use GitHub API to get latest release
    const apiUrl = `https://api.github.com/repos/${owner}/${repo}/releases/latest`;
    const response = await fetch(apiUrl, { headers: getHeaders() });
    
    if (response.status !== 200) {
      if (response.status === 403) {
        console.error(`\nGitHub API rate limit exceeded!`);
        console.error(`Unauthenticated: 60 requests/hour | Authenticated: 5,000 requests/hour`);
        console.error(`Set GITHUB_TOKEN environment variable to increase rate limit.`);
        process.exit(1);
      } else if (response.status === 404) {
        // No releases found, this is common and not an error
      } else {
        console.error(`GitHub API error ${response.status} for ${owner}/${repo}`);
      }
      return null;
    }
    
    const data = await response.json();
    
    if (data.tag_name) {
      // Get the commit SHA for the latest release tag
      const commit = await getCommitForRef(owner, repo, data.tag_name);
      if (commit) {
        return { tag: data.tag_name, commit };
      }
    }
    
    return null;
  } catch (e) {
    console.error(`Error fetching release for ${actionRef}:`, e instanceof Error ? e.message : e);
    return null;
  }
}

(async function main() {
  try {
    if (GITHUB_TOKEN) {
      console.log("Using authenticated GitHub API requests\n");
    } else {
      console.log("Using unauthenticated GitHub API requests (60/hour limit)");
      console.log("Set GITHUB_TOKEN environment variable for 5,000/hour limit");
      console.log("Create a token at: https://github.com/settings/tokens/new?description=check-outdated-actions&scopes=public_repo");
      console.log("(Only public_repo scope is needed for reading public repositories)\n");
    }
    
    const settings = require("./settings.json");
    const actionsByType = await findActionsInWorkflows(settings.folders);
    
    const jobUses = Array.from(new Set(actionsByType.get("JOB_USES") || [])).sort();
    const stepUses = Array.from(new Set(actionsByType.get("JOBS_STEP_USES") || [])).sort();
    
    // Process actions in batches for better performance
    const BATCH_SIZE = 10;
    let processed = 0;
    const total = stepUses.length + jobUses.length;
    
    const processAction = async (action: string): Promise<string> => {
      const [fullAction, currentVersion] = action.split("@");
      const parts = fullAction.split("/");
      
      if (parts.length < 2) {
        return action;
      }
      
      const owner = parts[0];
      const repo = parts[1];
      
      const latestRelease = await getLatestRelease(action);
      
      if (!latestRelease) {
        return action;
      }
      
      // Check if currentVersion is already a commit SHA (40 hex chars)
      const isCommitSHA = /^[0-9a-f]{40}$/i.test(currentVersion);
      
      let currentCommit: string | null;
      if (isCommitSHA) {
        // If it's already a commit SHA, use it directly
        currentCommit = currentVersion.toLowerCase();
      } else {
        // Otherwise, resolve the tag/ref to a commit
        currentCommit = await getCommitForRef(owner, repo, currentVersion);
      }
      
      if (!currentCommit) {
        return action;
      }
      
      // Compare commits - if they match, the current version resolves to the latest release
      if (currentCommit === latestRelease.commit) {
        return `${action} (latest)`;
      } else {
        return `${action} -> ${latestRelease.tag} ${latestRelease.commit}`;
      }
    };
    
    const processBatch = async (actions: string[]): Promise<string[]> => {
      const results: string[] = [];
      for (let i = 0; i < actions.length; i += BATCH_SIZE) {
        const batch = actions.slice(i, i + BATCH_SIZE);
        const batchResults = await Promise.all(batch.map(processAction));
        results.push(...batchResults);
        processed += batch.length;
        process.stderr.write(`\rProcessing... ${processed}/${total}`);
      }
      return results;
    };
    
    console.log("=== Step uses ===");
    const stepResults = await processBatch(stepUses);
    process.stderr.write("\r" + " ".repeat(50) + "\r"); // Clear progress line
    for (const result of stepResults) {
      console.log(result);
    }
    
    console.log("\n=== Job uses ===");
    const jobResults = await processBatch(jobUses);
    process.stderr.write("\r" + " ".repeat(50) + "\r"); // Clear progress line
    for (const result of jobResults) {
      console.log(result);
    }
  } catch (e) {
    console.error("Unhandled error while checking actions", e);
    process.exitCode = 1;
  }
})();
