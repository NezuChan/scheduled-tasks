import { Store } from "@sapphire/pieces";
import { cast, Constructor } from "@sapphire/utilities";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";
import { Listener } from "./Listener.js";

export class ListenerStore extends Store<Listener> {
    public constructor() {
        super(cast<Constructor<Listener>>(Listener), { name: "listeners" });
        this.registerPath(resolve(dirname(fileURLToPath(import.meta.url)), "..", "Listeners"));
    }
}
