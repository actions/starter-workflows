#!/usr/bin/env npx ts-node
import { promises as fs } from "fs";
import { safeLoad } from "js-yaml";
import { basename, extname, join } from "path";
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
  iconName: string;
  categories: string[];
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

async function checkWorkflows(folders: string[]): Promise<WorkflowWithErrors[]> {
  const result: WorkflowWithErrors[] = []
  const workflow_template_names = new Set()
  for (const folder of folders) {
    const dir = await fs.readdir(folder, {
      withFileTypes: true,
    });

    for (const e of dir) {
      if (e.isFile()) {
        const fileType = basename(e.name, extname(e.name))

        const workflowFilePath = join(folder, e.name);
        const propertiesFilePath = join(folder, "properties", `${fileType}.properties.json`)

        const workflowWithErrors = await checkWorkflow(workflowFilePath, propertiesFilePath);
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

async function checkWorkflow(workflowPath: string, propertiesPath: string): Promise<WorkflowWithErrors> {
  let workflowErrors: WorkflowWithErrors = {
    id: workflowPath,
    name: null,
    errors: []
  }

  try {
    const workflowFileContent = await fs.readFile(workflowPath, "utf8");
    safeLoad(workflowFileContent); // Validate yaml parses without error

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
        try {
          await fs.access(`../../icons/${properties.iconName}.svg`)
        } catch (e) {
          workflowErrors.errors.push(`No icon named ${properties.iconName} found`)
        }
      }
      else {
        let iconName = properties.iconName.match(/^octicon\s+(.*)/)
        if(!iconName || iconName[1].split(".")[0].length <= 0) {
          workflowErrors.errors.push(`No icon named ${properties.iconName} found`)
        }
      }
      
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
      settings.folders
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
