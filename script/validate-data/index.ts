#!/usr/bin/env npx ts-node
import { promises as fs } from "fs";
import { safeLoad } from "js-yaml";
import { basename, extname, join, dirname } from "path";
import { Validator as validator } from "jsonschema";
import { endGroup, error, info, setFailed, startGroup } from '@actions/core';

interface WorkflowWithErrors {
  id: string;
  name: string;
  errors: string[];
}

interface WorkflowProperties {
  name: string;
  description: string;
  creator: string;
  iconName: string;
  categories: string[];
}

const yamlWorkflowExtensions = [".yml", ".yaml"];

function getSupportedWorkflowExtensions(folder: string): string[] {
  if (basename(folder).toLowerCase() === "agentic") {
    return [...yamlWorkflowExtensions, ".md"];
  }

  return yamlWorkflowExtensions;
}

const propertiesSchema = {
  type: "object",
  properties: {
    name: { type: "string", required: true , "minLength": 1},
    description: { type: "string", required: true },
    creator: { type: "string", required: false },
    iconName: { type: "string", required: true },
    categories: {
      anyOf: [
        {
          type: "array",
          items: { type: "string" }
        },
        {
          type: "null",
        }
      ],
      required: true
    },
  }
}

async function checkWorkflows(folders: string[], allowed_categories: object[]): Promise<WorkflowWithErrors[]> {
  const result: WorkflowWithErrors[] = []
  const workflow_template_names = new Set()
  for (const folder of folders) {
    const supportedWorkflowExtensions = getSupportedWorkflowExtensions(folder);
    const dir = await fs.readdir(folder, {
      withFileTypes: true,
    });

    for (const e of dir) {
      if (e.isFile() && supportedWorkflowExtensions.includes(extname(e.name))) {
        const fileType = basename(e.name, extname(e.name))

        const workflowFilePath = join(folder, e.name);
        const propertiesFilePath = join(folder, "properties", `${fileType}.properties.json`)

        const workflowWithErrors = await checkWorkflow(workflowFilePath, propertiesFilePath, allowed_categories);
        if(workflowWithErrors.name && workflow_template_names.size == workflow_template_names.add(workflowWithErrors.name).size) {
          workflowWithErrors.errors.push(`Workflow template name "${workflowWithErrors.name}" already exists`) 
        }
        if (workflowWithErrors.errors.length > 0) {
          result.push(workflowWithErrors)
        }
      }
    }
  }

  return result;
}

function getMarkdownFrontmatter(workflowPath: string, workflowFileContent: string): string {
  const frontmatterMatch = workflowFileContent.match(/^---\r?\n([\s\S]*?)\r?\n---(?:\r?\n|$)/);

  if (!frontmatterMatch) {
    throw new Error(`Markdown workflow ${workflowPath} must start with valid YAML frontmatter`);
  }

  return frontmatterMatch[1];
}

function validateWorkflowContent(workflowPath: string, workflowFileContent: string): void {
  const extension = extname(workflowPath).toLowerCase();

  if (yamlWorkflowExtensions.includes(extension)) {
    safeLoad(workflowFileContent);
    return;
  }

  if (extension === ".md") {
    safeLoad(getMarkdownFrontmatter(workflowPath, workflowFileContent));
    return;
  }

  throw new Error(`Unsupported workflow extension ${extension}`);
}

async function checkWorkflow(workflowPath: string, propertiesPath: string, allowed_categories: object[]): Promise<WorkflowWithErrors> {
  let workflowErrors: WorkflowWithErrors = {
    id: workflowPath,
    name: null,
    errors: []
  }
  try {
    const workflowFileContent = await fs.readFile(workflowPath, "utf8");
    validateWorkflowContent(workflowPath, workflowFileContent);

    const propertiesFileContent = await fs.readFile(propertiesPath, "utf8")
    const properties: WorkflowProperties = JSON.parse(propertiesFileContent)
    if(properties.name && properties.name.trim().length > 0) {
      workflowErrors.name = properties.name
    }
    let v = new validator();
    const res = v.validate(properties, propertiesSchema)
    workflowErrors.errors = res.errors.map(e => e.toString())
    
    if (properties.iconName) {
      if(! /^octicon\s+/.test(properties.iconName)) {
        let svgIconName: string | undefined = properties.iconName
        if (properties.iconName.startsWith("lucide ")) {
          const lucideName = properties.iconName.slice("lucide ".length).split(".")[0].trim()
          if(!lucideName) {
            workflowErrors.errors.push(`No icon named ${properties.iconName} found`)
            svgIconName = undefined
          } else {
            svgIconName = `lucide-${lucideName}`
          }
        }

        if(!svgIconName) {
          workflowErrors.errors.push(`No icon named ${properties.iconName} found`)
        } else {
          try {
            await fs.access(`../../icons/${svgIconName}.svg`)
          } catch (e) {
            workflowErrors.errors.push(`No icon named ${properties.iconName} found`)
          }
        }
      }
      else {
        let iconName = properties.iconName.match(/^octicon\s+(.*)/)
        if(!iconName || iconName[1].split(".")[0].length <= 0) {
          workflowErrors.errors.push(`No icon named ${properties.iconName} found`)
        }
      }
      
    }
    var path = dirname(workflowPath)
    var folder_categories = allowed_categories.find( category => category["path"] == path)["categories"]
    if (!workflowPath.endsWith("blank.yml")) {
      if(!properties.categories || properties.categories.length == 0) {
        workflowErrors.errors.push(`Workflow categories cannot be null or empty`)
      } 
      else if(!folder_categories.some(category => properties.categories[0].toLowerCase() == category.toLowerCase())) {
        workflowErrors.errors.push(`The first category in properties.json categories for workflow in ${basename(path)} folder must be one of "${folder_categories}. Either move the workflow to an appropriate directory or change the category."`)
      }
    }

    if(basename(path).toLowerCase() == 'deployments' && !properties.creator) {
      workflowErrors.errors.push(`The "creator" in properties.json must be present.`)
    }
  } catch (e) {
    workflowErrors.errors.push(e.toString())
  }
  return workflowErrors;
}

(async function main() {
  try {
    const settings = require("./settings.json");
    const erroredWorkflows = await checkWorkflows(
      settings.folders, settings.allowed_categories
    )

    if (erroredWorkflows.length > 0) {
      startGroup(`😟 - Found ${erroredWorkflows.length} workflows with errors:`);
      erroredWorkflows.forEach(erroredWorkflow => {
        error(`Errors in ${erroredWorkflow.id} - ${erroredWorkflow.errors.map(e => e.toString()).join(", ")}`)
      })
      endGroup();
      setFailed(`Found ${erroredWorkflows.length} workflows with errors`);
    } else {
      info("🎉🤘 - Found no workflows with errors!")
    }
  } catch (e) {
    error(`Unhandled error while syncing workflows: ${e}`);
    setFailed(`Unhandled error`)
  }
})();
