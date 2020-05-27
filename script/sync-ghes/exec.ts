import { spawn } from "child_process";

export class ExecResult {
  stdout = "";
  exitCode = 0;
}

/**
 * Executes a process
 */
export async function exec(
  command: string,
  args: string[] = [],
  allowAllExitCodes: boolean = false
): Promise<ExecResult> {
  process.stdout.write(`EXEC: ${command} ${args.join(" ")}\n`);
  return new Promise((resolve, reject) => {
    const execResult = new ExecResult();
    const cp = spawn(command, args, {});

    // STDOUT
    cp.stdout.on("data", (data) => {
      process.stdout.write(data);
      execResult.stdout += data.toString();
    });

    // STDERR
    cp.stderr.on("data", (data) => {
      process.stderr.write(data);
    });

    // Close
    cp.on("close", (code) => {
      execResult.exitCode = code;
      if (code === 0 || allowAllExitCodes) {
        resolve(execResult);
      } else {
        reject(new Error(`Command exited with code ${code}`));
      }
    });
  });
}
