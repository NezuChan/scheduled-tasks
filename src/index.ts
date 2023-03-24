/* eslint-disable @typescript-eslint/no-loop-func */
/* eslint-disable @typescript-eslint/naming-convention */
/* eslint-disable @typescript-eslint/prefer-for-of */
import cluster from "node:cluster";
import { cpus } from "node:os";
import { TaskManager } from "./Structures/TaskManager.js";
import { Result } from "@sapphire/result";
import { createLogger } from "./Structures/Logger.js";

if (cluster.isPrimary) {
    const clusters = Number(process.env.TOTAL_CLUSTERS ?? cpus().length);
    const logger = createLogger("tasks", process.env.STORE_LOGS === "true", process.env.LOKI_HOST ? new URL(process.env.LOKI_HOST) : undefined);

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
