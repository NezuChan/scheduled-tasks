/* eslint-disable @typescript-eslint/restrict-template-expressions */
import { cast } from "@sapphire/utilities";
import Bull from "bull";
import { TaskManager } from "../Structures/TaskManager.js";

export async function handleJob(message: Record<string, any>, bull: Bull.Queue, clusterId: number, manager: TaskManager): Promise<string> {
    switch (message.type) {
        case "add": {
            if (!message.payload) {
                return JSON.stringify({ message: "Please provide payload to sent when job is done !" });
            }
            if (!message.options) {
                return JSON.stringify({ message: "Please provide options to be passed to bull job !" });
            }
            if (!message.name) {
                return JSON.stringify({ message: "Please provide name of the job to be created !" });
            }
            manager.logger.info(`Adding job ${message.name} to queue ${bull.name} in cluster ${manager.clusterId}`);
            const job = await bull.add(cast<string>(message.name), message.payload, cast<Bull.JobOptions>(message.options));
            return JSON.stringify({ message: "Job added successfully !", jobId: job.id.toString(), fromCluster: clusterId });
        }

        case "remove": {
            if (!message.jobId) {
                return JSON.stringify({ message: "Please provide jobId to be removed !" });
            }

            manager.logger.info(`Received request to remove job ${message.jobId} in cluster ${clusterId}`);

            const job = await bull.getJob(cast<string>(message.jobId));
            if (!job) {
                manager.logger.info(`Job with id ${message.jobId} not found in cluster ${clusterId}`);
                return JSON.stringify({ message: "Job not found !" });
            }

            await job.remove();
            manager.logger.info(`Job with id ${message.jobId} removed successfully !`);
            return JSON.stringify({ message: "Job success deleted !", fromCluster: clusterId });
        }
        default: {
            manager.logger.info(message, `Unhandled job type ${message.type}`);
            return JSON.stringify({ message: "Unhandled job, please open issue if you need this to be handled !" });
        }
    }
}
