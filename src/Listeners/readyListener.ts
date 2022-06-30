import { Listener, ListenerOptions } from "../Stores/Listener.js";
import { ApplyOptions } from "../Utilities/Decorators/ApplyOptions.js";

@ApplyOptions<ListenerOptions>({
    name: "ready"
})

export class readyListener extends Listener {
    public run(): void {
        this.container.manager.logger.info(`Task Manager cluster ${this.container.manager.clusterId} is ready.`);
    }
}
