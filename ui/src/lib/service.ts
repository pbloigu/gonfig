import { goto } from "$app/navigation";
import { PersistentState } from '@friendofsvelte/state';
const { MODE, VITE_API_URL } = import.meta.env;

import createClient, { type Middleware } from "openapi-fetch";
import type { paths, components } from "./client/api";

export interface Error {
    code?: number
    message?: string
}

export type ScriptParam = {
    key: string;
    value: string;
};

interface Token {
    value: string | null
}
export const token = new PersistentState<Token>("token", {
    value: null
}, "sessionStorage")

const resolveBaseUri = function (): string {
    if (MODE == "development") {
        return VITE_API_URL
    } else {
        return window.location.protocol + "//" + window.location.host
    }
}


const client = createClient<paths>({ baseUrl: resolveBaseUri() });

const middleware: Middleware = {
    async onRequest({ request, schemaPath }) {
        if (!schemaPath.startsWith("/login")) {
            request.headers.set("Authorization", "Bearer " + token.current.value);
        }
        return request;
    },
    async onResponse({ request, response, schemaPath }) {
        if (schemaPath.startsWith("/logout")) {
            token.current.value = null
        }
        if (!response.ok) {
            // it's okay, you can't always have what you want
            if (schemaPath.startsWith("/application/{id}/trigger/status")
                && request.method.toUpperCase() == "GET"
                && response.status == 404) {
                return;
            }
            // it's okay, this is a validation endpoint
            else if (schemaPath.startsWith("/cron/expression")
                && request.method.toUpperCase() == "POST"
                && response.status == 404) {
                return;
            } else if (response.status == 401) {
                goto("/login")
            } else {
                goto("/error", {
                    state: {
                        code: response.status,
                        message: response.text
                    }
                })
            }
        } else {
            return response;
        }

    },
    async onError({ error }) {
        // wrap errors thrown by fetch
        return new Error("Oops, fetch failed", { cause: error });
    },
};

client.use(middleware)

export const Login = async function (username: string, password: string) {
    const { data } = await client.POST("/login", {
        body: {
            password: password,
            username: username
        }
    })
    if (data) {
        token.current.value = data.token
        goto("/applications")
    }
}

export const Logout = async function () {
    const { data } = await client.GET("/logout")
    goto("/")
}

export const ListApplications = async function (): Promise<components["schemas"]["Application"][]> {
    const { data } = await client.GET("/applications")
    if (data) {
        return Promise.resolve(data)
    } else {
        return Promise.resolve([])
    }
}

export const GetApplication = async function (appId: string): Promise<components["schemas"]["Application"]> {
    const { data } = await client.GET("/application/{id}", {
        params: {
            path: {
                id: appId
            }
        }
    })
    if (data) {
        return Promise.resolve(data)
    } else {
        return Promise.resolve({} as components["schemas"]["Application"])
    }
}

export const AddApplication = async function (app: components["schemas"]["Application"]): Promise<components["schemas"]["Application"]> {
    const { data } = await client.POST("/application", {
        body: app
    })
    if (data) {
        return Promise.resolve(data)
    } else {
        return Promise.resolve({} as components["schemas"]["Application"])
    }
}

export const AddSeries = async function (name: string, appId: string): Promise<components["schemas"]["Series"]> {
    const { data } = await client.POST("/application/{id}/series", {
        params: {
            path: {
                id: appId
            }
        },
        body: {
            name: name
        }
    })
    if (data) {
        return Promise.resolve(data)
    } else {
        return Promise.resolve({} as components["schemas"]["Series"])
    }
}

export const UpdateConfiguration = async function (appId: string, configuration: string) {
    await client.POST("/application/{id}/configuration", {
        params: {
            path: {
                id: appId
            }
        },
        body: {
            data: configuration
        }
    })
}

export const DeleteApplication = async function (appId: string) {
    await client.DELETE("/application/{id}", {
        params: {
            path: {
                id: appId
            }
        }
    })
}

export const ListSeries = async function (appId: string): Promise<components["schemas"]["Series"][]> {
    const { data } = await client.GET("/application/{id}/series", {
        params: {
            path: {
                id: appId
            }
        }
    })
    if (data) {
        return Promise.resolve(data)
    } else {
        return Promise.resolve([])
    }
}


export const ListSeriesValues = async function name(name: string, appId: string, page: number): Promise<components["schemas"]["SeriesValues"]> {
    const { data } = await client.GET("/application/{id}/series/{name}/values", {
        params: {
            path: {
                id: appId,
                name: name
            }
        }
    })
    if (data) {
        return Promise.resolve(data)
    } else {
        return Promise.resolve({} as components["schemas"]["SeriesValues"])
    }
}

export const GetStatusChangeTrigger = async function (appId: string): Promise<components["schemas"]["StatusChangeTrigger"] | null> {
    const { data } = await client.GET("/application/{id}/trigger/status", {
        params: {
            path: {
                id: appId
            }
        }
    })
    if (data) {
        return Promise.resolve(data)
    } else {
        return Promise.resolve(null)
    }
}

export const AddStatusChangeTrigger = async function (appId: string, trigger: components["schemas"]["StatusChangeTrigger"]): Promise<components["schemas"]["StatusChangeTrigger"]> {
    const { data } = await client.POST("/application/{id}/trigger/status", {
        params: {
            path: {
                id: appId
            }
        },
        body: trigger
    })
    if (data) {
        return Promise.resolve(data)
    } else {
        return Promise.resolve({} as components["schemas"]["StatusChangeTrigger"])
    }
}

export const UpdateStatusChangeTrigger = async function (appId: string, trigger: components["schemas"]["StatusChangeTrigger"]) {
    await client.PUT("/application/{id}/trigger/status", {
        params: {
            path: {
                id: appId
            }
        },
        body: trigger
    })
}

export const DeleteStatusChangeTrigger = async function (appId: string) {
    await client.DELETE("/application/{id}/trigger/status", {
        params: {
            path: {
                id: appId
            }
        }
    })
}

export const IsValid = async function (cronExpression: components["schemas"]["CronValidationRequest"]): Promise<boolean> {
    const { data } = await client.POST("/cron/expression", {
        body: cronExpression
    })
    return data ? Promise.resolve(true) : Promise.resolve(false);
}


export const Execute = async function (script: string, params: ScriptParam[]): Promise<components["schemas"]["ScriptExecutionResponse"]> {
    let req: components["schemas"]["ScriptExecutionRequest"] = {
        script: script,
        context: {}
    }
    params.forEach(p => {
        req.context[p.key] = p.value
    })

    const { data } = await client.POST("/scripting/execute", {
        body: req
    })
    if (data) {
        return Promise.resolve(data)
    } else {
        return Promise.resolve({} as components["schemas"]["ScriptExecutionResponse"])
    }
}

