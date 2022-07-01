/* eslint-disable @typescript-eslint/no-loop-func */
/* eslint-disable @typescript-eslint/naming-convention */
/* eslint-disable @typescript-eslint/prefer-for-of */
import cluster from "node:cluster";
import { resolve } from "node:path";
import { cpus } from "node:os";
import pino from "pino";
import { Util } from "./Utilities/Util.js";
import { TaskManager } from "./Structures/TaskManager.js";
import { Result } from "@sapphire/result";

if (cluster.isPrimary) {
    const clusters = Number(process.env.TOTAL_CLUSTERS ?? cpus().length);
    const date = (): string => Util.formatDate(Intl.DateTimeFormat("en-US", {
        year: "numeric",
        month: "numeric",
        day: "numeric",
        hour12: false
    }));

    const logger = pino({
        name: "scheduled-tasks",
        timestamp: true,
        level: process.env.NODE_ENV === "production" ? "info" : "trace",
        formatters: {
            bindings: () => ({
                pid: "Scheduled Tasks"
            })
        },
        transport: {
            targets: [
                { target: "pino/file", level: "info", options: { destination: resolve(process.cwd(), "logs", `tasks-${date()}.log`) } },
                { target: "pino-pretty", level: process.env.NODE_ENV === "production" ? "info" : "trace", options: { translateTime: "SYS:yyyy-mm-dd HH:MM:ss.l o" } }
            ]
        }
    });


    logger.info(`Starting Scheduled Tasks in ${clusters} clusters`);
    for (let index = 0; index < clusters; index++) {
        logger.info(`Launching Scheduled Tasks cluster ${index}`);
        const clusterResult = Result.from(() => cluster.fork({ CLUSTER_ID: index, ...process.env }));
        if (clusterResult.isErr()) {
            logger.error(`Failed to launch Scheduled Tasks cluster ${index}`);
            continue;
        } else {
            logger.info(`Launched Scheduled Tasks cluster ${index}`);
        }
    }
} else {
    await new TaskManager().initialize();
}
