#!/usr/bin/env npx ts-node
import { promises as fs } from "fs";
import { safeLoad } from "js-yaml";
import { basename, extname, join } from "path";
import { exec } from "./exec";

interface WorkflowDesc {
  folder: string;
  id: string;
  iconName?: string;
  iconType?: "svg" | "octicon";
}

interface WorkflowProperties {
  name: string;

  description: string;

  iconName?: string;

  categories: string[] | null;

  creator?: string;

  enterprise?: boolean;
}

interface WorkflowsCheckResult {
  compatibleWorkflows: WorkflowDesc[];
  incompatibleWorkflows: WorkflowDesc[];
}

async function checkWorkflows(
  folders: string[],
  enabledActions: string[],
  partners: string[]
): Promise<WorkflowsCheckResult> {
  const result: WorkflowsCheckResult = {
    compatibleWorkflows: [],
    incompatibleWorkflows: [],
  };
  const partnersSet = new Set(partners.map((x) => x.toLowerCase()));

  for (const folder of folders) {
    const dir = await fs.readdir(folder, {
      withFileTypes: true,
    });

    for (const e of dir) {
      if (e.isFile() && extname(e.name) === ".yml") {
        const workflowFilePath = join(folder, e.name);
        const workflowId = basename(e.name, extname(e.name));
        const workflowProperties: WorkflowProperties = require(join(
          folder,
          "properties",
          `${workflowId}.properties.json`
        ));
        const iconName: string | undefined = workflowProperties["iconName"];

        const isPartnerWorkflow = workflowProperties.creator ? partnersSet.has(workflowProperties.creator.toLowerCase()) : false;

        const enabled =
          !isPartnerWorkflow &&
          (workflowProperties.enterprise === true || basename(folder) !== 'code-scanning') &&
          (await checkWorkflow(workflowFilePath, enabledActions));

        const workflowDesc: WorkflowDesc = {
          folder,
          id: workflowId,
          iconName,
          iconType:
            iconName && iconName.startsWith("octicon") ? "octicon" : "svg",
        };

        if (!enabled) {
          result.incompatibleWorkflows.push(workflowDesc);
        } else {
          result.compatibleWorkflows.push(workflowDesc);
        }
      }
    }
  }

  return result;
}

/**
 * Check if a workflow uses only the given set of actions.
 *
 * @param workflowPath Path to workflow yaml file
 * @param enabledActions List of enabled actions
 */
async function checkWorkflow(
  workflowPath: string,
  enabledActions: string[]
): Promise<boolean> {
  // Create set with lowercase action names for easier, case-insensitive lookup
  const enabledActionsSet = new Set(enabledActions.map((x) => x.toLowerCase()));
  try {
    const workflowFileContent = await fs.readFile(workflowPath, "utf8");
    const workflow = safeLoad(workflowFileContent);

    for (const job of Object.keys(workflow.jobs || {}).map(
      (k) => workflow.jobs[k]
    )) {
      for (const step of job.steps || []) {
        if (!!step.uses) {
          // Check if allowed action
          const [actionName, _] = step.uses.split("@");
          const actionNwo = actionName.split("/").slice(0, 2).join("/");
          if (!enabledActionsSet.has(actionNwo.toLowerCase())) {
            console.info(
              `Workflow ${workflowPath} uses '${actionName}' which is not supported for GHES.`
            );
            return false;
          }
        }
      }
    }

    // All used actions are enabled 🎉
    return true;
  } catch (e) {
    console.error("Error while checking workflow", e);
    throw e;
  }
}

(async function main() {
  try {
    const settings = require("./settings.json");

    const result = await checkWorkflows(
      settings.folders,
      settings.enabledActions,
      settings.partners
    );

    console.group(
      `Found ${result.compatibleWorkflows.length} starter workflows compatible with GHES:`
    );
    console.log(
      result.compatibleWorkflows.map((x) => `${x.folder}/${x.id}`).join("\n")
    );
    console.groupEnd();

    console.group(
      `Ignored ${result.incompatibleWorkflows.length} starter-workflows incompatible with GHES:`
    );
    console.log(
      result.incompatibleWorkflows.map((x) => `${x.folder}/${x.id}`).join("\n")
    );
    console.groupEnd();

    console.log("Switch to GHES branch");
    await exec("git", ["checkout", "ghes"]);

    // In order to sync from main, we might need to remove some workflows, add some
    // and modify others. The lazy approach is to delete all workflows first (except from read-only folders), and then
    // just bring the compatible ones over from the main branch. We let git figure out
    // whether it's a deletion, add, or modify and commit the new state.
    console.log("Remove all workflows");
    await exec("rm", ["-fr", ...settings.folders]);
    await exec("rm", ["-fr", "../../icons"]);

    // Bring back the read-only folders
    console.log("Restore read-only folders");
    for (let i = 0; i < settings.readOnlyFolders.length; i++) {
      await exec("git", [
        "checkout",
        settings.readOnlyFolders[i]
      ]);
    }

    console.log("Sync changes from main for compatible workflows");
    await exec("git", [
      "checkout",
      "main",
      "--",
      ...Array.prototype.concat.apply(
        [],
        result.compatibleWorkflows.map((x) => {
          const r = [];

          // Don't touch read-only folders
          if (!settings.readOnlyFolders.includes(x.folder)) {
            r.push(join(x.folder, `${x.id}.yml`));
            r.push(join(x.folder, "properties", `${x.id}.properties.json`));
          };

          if (x.iconType === "svg") {
            r.push(join("../../icons", `${x.iconName}.svg`));
          }

          return r;
        })
      ),
    ]);

    // The v4 versions of upload and download artifact are not yet supported on GHES
    console.group("Updating all compatible workflows to use v3 of the artifact actions");
    for (const workflow of result.compatibleWorkflows) {
      const path = join(workflow.folder, `${workflow.id}.yml`);
      console.log(`Updating ${path}`);
      const contents = await fs.readFile(path, "utf8");

      if (contents.includes("actions/upload-artifact@v4") || contents.includes("actions/download-artifact@v4")) {
        console.log("Found v4 artifact actions, updating to v3");
      } else {
        continue;
      }

      let updatedContents = contents.replace(/actions\/upload-artifact@v4/g, "actions/upload-artifact@v3");
      updatedContents = updatedContents.replace(/actions\/download-artifact@v4/g, "actions/download-artifact@v3");

      await fs.writeFile(path, updatedContents);
    }
    console.groupEnd();

  } catch (e) {
    console.error("Unhandled error while syncing workflows", e);
    process.exitCode = 1;
  }
})();
