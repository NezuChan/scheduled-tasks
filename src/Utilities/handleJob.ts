/* eslint-disable @typescript-eslint/restrict-template-expressions */
import { cast } from "@sapphire/utilities";
import Bull from "bull";
import { TaskManager } from "../Structures/TaskManager.js";

export async function handleJob(message: Record<string, any>, bull: Bull.Queue, clusterId: number, manager: TaskManager): Promise<string> {
    switch (message.type) {
        case "add": {
            if (!message.data) {
                return JSON.stringify({ message: "Please provide payload to sent when job is done !" });
            }
            if (!message.options) {
                return JSON.stringify({ message: "Please provide options to be passed to bull job !" });
            }
            if (!message.name) {
                return JSON.stringify({ message: "Please provide name of the job to be created !" });
            }
            manager.logger.info(`Adding job ${message.name} to queue ${bull.name} in cluster ${manager.clusterId}`);
            const job = await bull.add(cast<string>(message.name), message.data, cast<Bull.JobOptions>(message.options));
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
        case "pause": {
            manager.logger.info(`Received request to pause queue ${bull.name} in cluster ${clusterId}`);
            await bull.pause(cast<boolean | undefined>(message.isLocal), cast<boolean | undefined>(message.doNotWaitActive));
            return JSON.stringify({ message: "Bull job success paused !", fromCluster: clusterId });
        }
        case "resume": {
            manager.logger.info(`Received request to resume queue ${bull.name} in cluster ${clusterId}`);
            await bull.resume(cast<boolean | undefined>(message.isLocal));
            return JSON.stringify({ message: "Bull job success resumed !", fromCluster: clusterId });
        }
        case "getJob": {
            if (!message.jobId) {
                return JSON.stringify({ message: "Please provide jobId to lookup !" });
            }

            manager.logger.info(`Received request to get job ${message.jobId} in cluster ${clusterId}`);

            const job = await bull.getJob(cast<string>(message.jobId));
            if (!job) {
                manager.logger.info(`Job with id ${message.jobId} not found in cluster ${clusterId}`);
                return JSON.stringify({ message: "Job not found !" });
            }

            return JSON.stringify({ message: "Job found !", job: job.toJSON(), fromCluster: clusterId });
        }
        case "getJobs": {
            manager.logger.info(`Received request to get jobs in cluster ${clusterId}`);

            if (!message.types) {
                return JSON.stringify({ message: "Please provide types to lookup !" });
            }

            const jobs = await bull.getJobs(cast<Bull.JobStatus[]>(message.types), cast<number>(message.start), cast<number>(message.end), cast<boolean>(message.asc));
            return JSON.stringify({ message: "Jobs found !", jobs: jobs.map(j => j.toJSON()), fromCluster: clusterId });
        }
        case "getWaiting": {
            manager.logger.info(`Received request to get waiting jobs in cluster ${clusterId}`);

            const jobs = await bull.getWaiting();
            return JSON.stringify({ message: "Jobs found !", jobs: jobs.map(j => j.toJSON()), fromCluster: clusterId });
        }
        case "getActive": {
            manager.logger.info(`Received request to get active jobs in cluster ${clusterId}`);

            const jobs = await bull.getActive();
            return JSON.stringify({ message: "Jobs found !", jobs: jobs.map(j => j.toJSON()), fromCluster: clusterId });
        }
        case "getCompleted": {
            manager.logger.info(`Received request to get completed jobs in cluster ${clusterId}`);

            const jobs = await bull.getCompleted();
            return JSON.stringify({ message: "Jobs found !", jobs: jobs.map(j => j.toJSON()), fromCluster: clusterId });
        }
        case "getFailed": {
            manager.logger.info(`Received request to get failed jobs in cluster ${clusterId}`);

            const jobs = await bull.getFailed();
            return JSON.stringify({ message: "Jobs found !", jobs: jobs.map(j => j.toJSON()), fromCluster: clusterId });
        }
        case "getDelayed": {
            manager.logger.info(`Received request to get delayed jobs in cluster ${clusterId}`);

            const jobs = await bull.getDelayed();
            return JSON.stringify({ message: "Jobs found !", jobs: jobs.map(j => j.toJSON()), fromCluster: clusterId });
        }
        case "getActiveCount": {
            manager.logger.info(`Received request to get active count jobs in cluster ${clusterId}`);

            const active = await bull.getActiveCount();
            return JSON.stringify({ message: "Jobs found !", active, fromCluster: clusterId });
        }
        case "getWaitingCount": {
            manager.logger.info(`Received request to get waiting jobs in cluster ${clusterId}`);

            const waiting = await bull.getWaitingCount();
            return JSON.stringify({ message: "Jobs found !", waiting, fromCluster: clusterId });
        }
        case "getCompletedCount": {
            manager.logger.info(`Received request to get completed jobs in cluster ${clusterId}`);

            const completed = await bull.getCompletedCount();
            return JSON.stringify({ message: "Jobs found !", completed, fromCluster: clusterId });
        }
        case "getFailedCount": {
            manager.logger.info(`Received request to get failed jobs in cluster ${clusterId}`);

            const failed = await bull.getFailedCount();
            return JSON.stringify({ message: "Jobs found !", failed, fromCluster: clusterId });
        }
        case "getDelayedCount": {
            manager.logger.info(`Received request to get delayed jobs in cluster ${clusterId}`);

            const delayed = await bull.getDelayedCount();
            return JSON.stringify({ message: "Jobs found !", delayed, fromCluster: clusterId });
        }
        case "addBulk": {
            manager.logger.info(`Received request to add bulk jobs in cluster ${clusterId}`);
            if (!Array.isArray(message.jobs)) {
                return JSON.stringify({ message: "Please provide jobs to add !" });
            }
            const jobs = await bull.addBulk(cast<BulkJobType[]>(message.jobs));
            return JSON.stringify({ message: "Jobs added !", jobs: jobs.map(j => j.toJSON()), fromCluster: clusterId });
        }
        case "removeJobs": {
            manager.logger.info(`Received request to remove jobs in cluster ${clusterId}`);

            if (!message.pattern) {
                return JSON.stringify({ message: "Please provide jobIds to remove !" });
            }

            await bull.removeJobs(cast<string>(message.pattern));
            return JSON.stringify({ message: "Jobs removed !", fromCluster: clusterId });
        }
        case "getNextJob": {
            manager.logger.info(`Received request to get next job in cluster ${clusterId}`);

            const job = await bull.getNextJob();

            if (!job) {
                manager.logger.info(`No job found in cluster ${clusterId}`);
                return JSON.stringify({ message: "No job found !", fromCluster: clusterId });
            }

            return JSON.stringify({ message: "Job found !", job: job.toJSON(), fromCluster: clusterId });
        }
        case "getJobCountByTypes": {
            manager.logger.info(`Received request to get job count by types in cluster ${clusterId}`);

            if (!message.types) {
                return JSON.stringify({ message: "Please provide types to lookup !" });
            }

            const jobCountByTypes = await bull.getJobCountByTypes(cast<Bull.JobStatus | Bull.JobStatus[]>(message.types));
            return JSON.stringify({ message: "Job count by types found !", jobCountByTypes, fromCluster: clusterId });
        }
        case "nextRepeatableJob": {
            manager.logger.info(`Received request to get next repeatable job in cluster ${clusterId}`);

            if (!message.data) {
                return JSON.stringify({ message: "Please provide data to loop up !" });
            }
            if (!message.options) {
                return JSON.stringify({ message: "Please provide options to be passed to bull job !" });
            }
            if (!message.name) {
                return JSON.stringify({ message: "Please provide name of the job to look up !" });
            }
            const job = await bull.nextRepeatableJob(cast<string>(message.name), message.data, cast<Bull.JobOptions>(message.options));
            return JSON.stringify({ message: "Job found !", job: job.toJSON(), fromCluster: clusterId });
        }
        case "getRepeatableJobs": {
            manager.logger.info(`Received request to get repeatable jobs in cluster ${clusterId}`);

            const jobs = await bull.getRepeatableJobs(cast<number>(message.start), cast<number>(message.end), cast<boolean>(message.asc));
            return JSON.stringify({ message: "Jobs found !", jobs, fromCluster: clusterId });
        }
        default: {
            if (message.type) {
                manager.logger.warn(message, `Unhandled job type ${message.type}`);
                return JSON.stringify({ message: "Unhandled job, please open issue if you need this to be handled  or this is missing from the implementation !" });
            }
        }
    }
}

interface BulkJobType {
    name?: string | undefined;
    data: any;
    opts?: Omit<Bull.JobOptions, "repeat"> | undefined;
}
