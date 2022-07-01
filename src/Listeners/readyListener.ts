import { Listener, ListenerOptions } from "../Stores/Listener.js";
import { ApplyOptions } from "../Utilities/Decorators/ApplyOptions.js";

@ApplyOptions<ListenerOptions>({
    name: "ready"
})

export class readyListener extends Listener {
    public async run(): Promise<void> {
        const jobList = await this.container.manager.bull.getJobs(["delayed", "waiting"]);
        const jobRepeatedList = await this.container.manager.bull.getRepeatableJobs(0, -1);
        this.logger.info(`Scheduled Tasks cluster ${this.container.manager.clusterId} is ready. ${jobList.length} job(s), ${jobRepeatedList.length} repeatable job(s).`);
    }
}
