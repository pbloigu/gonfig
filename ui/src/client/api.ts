/* tslint:disable */
/* eslint-disable */

import axios, { type AxiosInstance, type AxiosRequestConfig } from "axios";
import type { Application, Configuration, Measurement, MeasurementValues, LoginRequest, LoginResponse, WithoutReadonly, WithoutWriteonly } from "./definitions";

export * from "./definitions";

export default class {
    public axios: AxiosInstance;

    constructor(configOrInstance: AxiosRequestConfig | AxiosInstance) {
        this.axios = 'interceptors' in configOrInstance
            ? configOrInstance
            : axios.create(configOrInstance)
    }

    private application_post(params: Record<string, never>, data: WithoutReadonly<Application>, options?: AxiosRequestConfig) {
        return this.axios.post<WithoutWriteonly<Application>>(
            "/application", data, options
        );
    }

    private id_get(params: {
        'id': string
    }, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<Application>>(
            "/application/{id}".replace(/{id}/, String(params["id"])), options
        );
    }

    private id_patch(params: {
        'id': string
    }, data: WithoutReadonly<Application>, options?: AxiosRequestConfig) {
        return this.axios.patch<WithoutWriteonly<Application>>(
            "/application/{id}".replace(/{id}/, String(params["id"])), data, options
        );
    }

    private id_delete(params: {
        'id': string
    }, options?: AxiosRequestConfig) {
        return this.axios.delete(
            "/application/{id}".replace(/{id}/, String(params["id"])), options
        );
    }

    private configuration_post(params: {
        'id': string
    }, data: WithoutReadonly<Configuration>, options?: AxiosRequestConfig) {
        return this.axios.post(
            "/application/{id}/configuration".replace(/{id}/, String(params["id"])), data, options
        );
    }

    private measurements_get(params: {
        'id': string
    }, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<Measurement>[]>(
            "/application/{id}/measurements".replace(/{id}/, String(params["id"])), options
        );
    }

    private name_get(params: {
        'id': string,
        'name': string
    }, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<MeasurementValues>>(
            "/application/{id}/measurements/{name}".replace(/{id}/, String(params["id"])).replace(/{name}/, String(params["name"])), options
        );
    }

    private applications_get(params: Record<string, never>, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<Application>[]>(
            "/applications", options
        );
    }

    private login_post(params: Record<string, never>, data: WithoutReadonly<LoginRequest>, options?: AxiosRequestConfig) {
        return this.axios.post<WithoutWriteonly<LoginResponse>>(
            "/login", data, options
        );
    }

    private logout_get(params: {
        'Authorization': string
    }, options?: AxiosRequestConfig) {
        return this.axios.get(
            "/logout",
            Object.assign(
                {},
                {
                    headers: pick(params, "Authorization"),
                },
                options
            )
        );
    }

    get Default() {
        return {
            application_post: this.application_post.bind(this),
            id_get: this.id_get.bind(this),
            id_patch: this.id_patch.bind(this),
            id_delete: this.id_delete.bind(this),
            configuration_post: this.configuration_post.bind(this),
            measurements_get: this.measurements_get.bind(this),
            name_get: this.name_get.bind(this),
            applications_get: this.applications_get.bind(this),
            login_post: this.login_post.bind(this),
            logout_get: this.logout_get.bind(this)
        };
    }
}

function pick<T, K extends keyof T>(obj: T, ...keys: K[]): Pick<T, K> {
    const ret: Pick<T, K> = {} as Pick<T, K>;
    keys.forEach(key => {
        if (obj && Object.keys(obj).includes(key as string))
            ret[key] = obj[key];
    });
    return ret;
}
