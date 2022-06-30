/* eslint-disable @typescript-eslint/no-unsafe-call */
/* eslint-disable @typescript-eslint/no-unsafe-member-access */
/* eslint-disable @typescript-eslint/no-shadow */
import { Piece, container, Container } from "@sapphire/pieces";
import { Ctor } from "@sapphire/utilities";

export function createProxy<T extends object>(target: T, handler: Omit<ProxyHandler<T>, "get">): T {
    return new Proxy(target, {
        ...handler,
        get: (target, property) => {
            const value = Reflect.get(target, property);
            return typeof value === "function" ? (...args: readonly unknown[]) => value.apply(target, args) : value;
        }
    });
}

export function createClassDecorator<TFunction extends (...args: any[]) => void>(fn: TFunction): ClassDecorator {
    return fn;
}

export function ApplyOptions<T extends Piece.Options>(optionsOrFn: T | ((parameters: ApplyOptionsCallbackParameters) => T)): ClassDecorator {
    return createClassDecorator((target: Ctor<ConstructorParameters<typeof Piece>, Piece>) => createProxy(target, {
        construct: (ctor, [context, baseOptions = {}]: [Piece.Context, Piece.Options]) => new ctor(context, {
            ...baseOptions,
            ...typeof optionsOrFn === "function" ? optionsOrFn({ container, context }) : optionsOrFn
        })
    }));
}

export interface ApplyOptionsCallbackParameters {
    container: Container;
    context: Piece.Context;
}
