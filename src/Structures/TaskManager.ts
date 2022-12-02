/* eslint-disable @typescript-eslint/restrict-template-expressions */
/* eslint-disable @typescript-eslint/no-base-to-string */
/* eslint-disable class-methods-use-this */
import { container, Piece, Store, StoreRegistry } from "@sapphire/pieces";
import Bull from "bull";
import EventEmitter from "node:events";
import { resolve } from "node:path";
import pino from "pino";
import { ListenerStore } from "../Stores/ListenerStore.js";
import { Util } from "../Utilities/Util.js";
import { createAmqp, RoutingPublisher, RpcSubscriber } from "@nezuchan/cordis-brokers";
import { handleJob } from "../Utilities/handleJob.js";
import { Result } from "@sapphire/result";
import { cast } from "@sapphire/utilities";

export class TaskManager extends EventEmitter {
    public stores = new StoreRegistry();
    public clusterId = parseInt(process.env.CLUSTER_ID!);
    public amqpSender!: RoutingPublisher<string, Record<string, any>>;
    public amqpReceiver!: RpcSubscriber<string, Record<string, any>>;
    public amqpReceiverCluster!: RpcSubscriber<string, Record<string, any>>;

    public bull = new Bull(`${process.env.QUEUE_NAME ?? "scheduled-tasks"}-cluster-${this.clusterId}`, {
        redis: {
            host: process.env.REDIS_HOST!,
            port: parseInt(process.env.REDIS_PORT!),
            username: process.env.REDIS_USERNAME,
            password: process.env.REDIS_PASSWORD,
            db: parseInt(process.env.REDIS_DB!)
        },
        defaultJobOptions: {
            removeOnComplete: true,
            removeOnFail: true,
            attempts: 2
        }
    });

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

    public async initialize(): Promise<void> {
        container.manager = this;
        this.logger.info(`Initializing Scheduled Tasks cluster ${this.clusterId}`);
        void this.bull.process("*", job => this.amqpSender.publish(job.name, typeof job.data === "object" ? job.data : JSON.stringify(job.data)));
        const { channel } = await createAmqp(process.env.AMQP_HOST!);
        this.amqpSender = new RoutingPublisher(channel);
        this.amqpReceiver = new RpcSubscriber(channel);
        this.amqpReceiverCluster = new RpcSubscriber(channel);
        await this.amqpReceiver.init({
            name: `${process.env.AMQP_QUEUE_NAME ?? "scheduled-tasks"}.send`,
            cb: async message => {
                const isJobReady = await this.bull.isReady();
                const result = await Result.fromAsync(() => handleJob(message, isJobReady, this.clusterId, this));
                if (result.isErr()) {
                    this.logger.error(result.unwrapErr());
                    return JSON.stringify({ error: cast<string>(result.unwrapErr()).toString() });
                }
                return result.unwrap();
            }
        });
        await this.amqpReceiverCluster.init({
            name: `${process.env.AMQP_QUEUE_NAME ?? "scheduled-tasks"}.send-cluster-${this.clusterId}`,
            cb: async message => {
                const isJobReady = await this.bull.isReady();
                const result = await Result.fromAsync(() => handleJob(message, isJobReady, this.clusterId, this));
                if (result.isErr()) {
                    this.logger.error(result.unwrapErr());
                    return JSON.stringify({ error: cast<string>(result.unwrapErr()).toString() });
                }
                return result.unwrap();
            }
        });
        await this.amqpSender.init({ name: `${process.env.AMQP_QUEUE_NAME ?? "scheduled-tasks"}.recv`, durable: true, exchangeType: "topic", useExchangeBinding: true });
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
