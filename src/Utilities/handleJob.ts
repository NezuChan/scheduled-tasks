import { cast } from "@sapphire/utilities";
import Bull from "bull";

export async function handleJob(message: Record<string, any>, bull: Bull.Queue, clusterId: number): Promise<string> {
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

            console.log(message.options);

            const job = await bull.add(cast<string>(message.name), message.payload, cast<Bull.JobOptions>(message.options));
            return JSON.stringify({ message: "Job added successfully !", jobId: job.id.toString(), fromCluster: clusterId });
        }

        case "remove": {
            if (!message.jobId) {
                return JSON.stringify({ message: "Please provide jobId to be removed !" });
            }

            const job = await bull.getJob(cast<string>(message.jobId));
            if (!job) {
                return JSON.stringify({ message: "Job not found !" });
            }

            await job.remove();
            return JSON.stringify({ message: "Job success deleted !", fromCluster: clusterId });
        }
        default: {
            return JSON.stringify({ message: "Unhandled job, please open issue if you need this to be handled !" });
        }
    }
}
