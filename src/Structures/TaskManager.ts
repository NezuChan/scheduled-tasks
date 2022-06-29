/* eslint-disable class-methods-use-this */
import { StoreRegistry } from "@sapphire/pieces";
import { resolve } from "node:path";
import pino from "pino";
import { Util } from "../Utilities/Util.js";

export class TaskManager {
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

    public initialize(clusterId: number): void {
        this.clusterId = clusterId;
        this.logger.info(`Initializing Scheduled Tasks cluster ${this.clusterId}`);
    }
}
