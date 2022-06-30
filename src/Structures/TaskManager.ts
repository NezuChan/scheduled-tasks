/* eslint-disable class-methods-use-this */
import { container, Piece, Store, StoreRegistry } from "@sapphire/pieces";
import EventEmitter from "node:events";
import { resolve } from "node:path";
import pino from "pino";
import { ListenerStore } from "../Stores/ListenerStore.js";
import { Util } from "../Utilities/Util.js";

export class TaskManager extends EventEmitter {
    public stores = new StoreRegistry();
    public clusterId!: number;
    public logger = pino({
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
                { target: "pino/file", level: "info", options: { destination: resolve(process.cwd(), "logs", `tasks-${this.date()}.log`) } },
                { target: "pino-pretty", level: process.env.NODE_ENV === "production" ? "info" : "trace", options: { translateTime: "SYS:yyyy-mm-dd HH:MM:ss.l o" } }
            ]
        }
    });

    public date(): string {
        return Util.formatDate(Intl.DateTimeFormat("en-US", {
            year: "numeric",
            month: "numeric",
            day: "numeric",
            hour12: false
        }));
    }

    public async initialize(clusterId: number): Promise<void> {
        this.clusterId = clusterId;
        container.manager = this;
        this.logger.info(`Initializing Scheduled Tasks cluster ${this.clusterId}`);
        this.stores.register(new ListenerStore());
        await Promise.all([...this.stores.values()].map((store: Store<Piece>) => store.loadAll()));
        this.emit("ready", this);
    }
}

declare module "@sapphire/pieces" {
    interface Container {
        manager: TaskManager;
    }
}
