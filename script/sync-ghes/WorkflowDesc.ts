import { promises as fs } from "fs";
import { safeLoad } from "js-yaml";
import { basename, extname, join } from "path";
import { exec } from "./exec";

export interface WorkflowDesc {
  readonly folder: string;
  readonly id: string;
  readonly iconName?: string;
  readonly iconType?: "svg" | "octicon";
}

export interface WorkflowsCheckResult {
  readonly compatibleWorkflows: ReadonlyArray<WorkflowDesc>;
  readonly incompatibleWorkflows: ReadonlyArray<WorkflowDesc>;
  /**
   * Returns true if all workflows are compatible with GHES.
   */
  get allCompatible(): boolean {
    return this.incompatibleWorkflows.length === 0;
  }
}

/**
 * Checks if a workflow is compatible with GHES.
 *
 * @param {string} workflowPath Path to workflow yaml file
 * @param {string[]} enabledActions List of enabled actions
 * @returns {Promise<boolean>} A promise that resolves to true if all actions used in the workflow are enabled, false otherwise.
 */
export async function checkWorkflow(
  workflowPath: string,
  enabledActions: string[]
): Promise<boolean> {
  /**
   * A set with lowercase action names for easier, case-insensitive lookup.
   * We use this set to check if an action is enabled for GHES.
   */
/**
 * A set of normalized (lowercase) action names that are enabled.
 * Used for efficient lookup of enabled actions with case-insensitive matching.
 */
  const enabledActionsSet = new Set(
    enabledActions.map((x) => x.toLowerCase())
  );

  try {
    const workflowFileContent = await fs.readFile(workflowPath, "utf8");
    const workflow = safeLoad(workflowFileContent);

    // Check each job in the workflow
    for (const jobName of Object.keys(workflow.jobs || {})) {
      const job = workflow.jobs[jobName];

      // Check each step in the job
      for (const step of job.steps || []) {
        if (!!step.uses) {
          // Split the uses string into action name and version
          const [actionName, _] = step.uses.split("@");

          // Check if the action is enabled
          if (!enabledActionsSet.has(actionName.toLowerCase())) {
            console.info(
              `Workflow ${workflowPath} uses '${actionName}' which is not supported for GHES.`
            );
            return false;
          }
        }
      }
    }

    // All used actions are enabled
    return true;
  } catch (e) {
    console.error("Error while checking workflow", e);
    throw e;
  }
// ( async function main ()
// {
//   try
//   {
//     const settings = require( "./settings.json" );

//     const result = await checkWorkflows(
//       settings.folders,
//       settings.enabledActions
//     );

//     console.group(
//       `Found ${ result.compatibleWorkflows.length } starter workflows compatible with GHES:`
//     );
//     console.log(
//       result.compatibleWorkflows.map( ( x ) => `${ x.folder }/${ x.id }` ).join( "\n" )
//     );
//     console.groupEnd();

//     console.group(
//       `Ignored ${ result.incompatibleWorkflows.length } starter-workflows incompatible with GHES:`
//     );
//     console.log(
//       result.incompatibleWorkflows.map( ( x ) => `${ x.folder }/${ x.id }` ).join( "\n" )
//     );
//     console.groupEnd();

//     console.log( "Switch to GHES branch" );
//     await exec( "git", ["checkout", "ghes"] );

//     // In order to sync from master, we might need to remove some workflows, add some
//     // and modify others. The lazy approach is to delete all workflows first, and then
//     // just bring the compatible ones over from the master branch. We let git figure out
//     // whether it's a deletion, add, or modify and commit the new state.
//     console.log( "Remove all workflows" );
//     await exec( "rm", ["-fr", ...settings.folders] );
//     await exec( "rm", ["-fr", "../../icons"] );

//     console.log( "Sync changes from master for compatible workflows" );
//     await exec( "git", [
//       "checkout",
//       "master",
//       "--",
//       ...Array.prototype.concat.apply(
//         [],
//         result.compatibleWorkflows.map( ( x ) =>
//         {
//           const r = [
//             join( x.folder, `${ x.id }.yml` ),
//             join( x.folder, "properties", `${ x.id }.properties.json` ),
//           ];

//           if ( x.iconType === "svg" )
//           {
//             r.push( join( "../../icons", `${ x.iconName }.svg` ) );
//           }

//           return r;
//         } )
//       ),
//     ] );
//   } catch ( e )
//   {
//     console.error( "Unhandled error while syncing workflows", e );
//     process.exitCode = 1;
//   }
// } )();
/**
 * Checks all workflows in the given folders and returns an object with two arrays.
 * The first array contains all workflows that are compatible with the given enabled actions,
 * and the second array contains all workflows that are not compatible.
 *
 * @param {string[]} folders The folders to search for workflows.
 * @param {string[]} enabledActions The enabled actions in GitHub Actions.
 * @returns {Promise<WorkflowsCheckResult>} An object with two arrays, one for compatible workflows and one for incompatible workflows.
 */
export async function checkWorkflows(
  folders: string[],
  enabledActions: string[]
): Promise<WorkflowsCheckResult> {
  const compatibleWorkflows: WorkflowDesc[] = [];
  const incompatibleWorkflows: WorkflowDesc[] = [];

  for (const folder of folders) {
    const files = await fs.readdir(folder);

    for (const file of files) {
      if (extname(file) === ".yml") {
        const id = basename(file, ".yml");
        const workflowPath = join(folder, file);

        const isCompatible = await checkWorkflow(workflowPath, enabledActions);

        const workflowDesc: WorkflowDesc = { folder, id };

        if (isCompatible) {
          compatibleWorkflows.push(workflowDesc);
        } else {
          incompatibleWorkflows.push(workflowDesc);
        }
      }
    }
  }

  return {
    compatibleWorkflows,
    incompatibleWorkflows,
  };
}
