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

    private addApplication(params: Record<string, never>, data: WithoutReadonly<Application>, options?: AxiosRequestConfig) {
        return this.axios.post<WithoutWriteonly<Application>>(
            "/application", data, options
        );
    }

    private getApplication(params: {
        'id': string
    }, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<Application>>(
            "/application/{id}".replace(/{id}/, String(params["id"])), options
        );
    }

    private updateApplication(params: {
        'id': string
    }, data: WithoutReadonly<Application>, options?: AxiosRequestConfig) {
        return this.axios.patch<WithoutWriteonly<Application>>(
            "/application/{id}".replace(/{id}/, String(params["id"])), data, options
        );
    }

    private deleteApplication(params: {
        'id': string
    }, options?: AxiosRequestConfig) {
        return this.axios.delete(
            "/application/{id}".replace(/{id}/, String(params["id"])), options
        );
    }

    private addConfiguration(params: {
        'id': string
    }, data: WithoutReadonly<Configuration>, options?: AxiosRequestConfig) {
        return this.axios.post(
            "/application/{id}/configuration".replace(/{id}/, String(params["id"])), data, options
        );
    }

    private addMeasurement(params: {
        'id': string
    }, data: WithoutReadonly<Measurement>, options?: AxiosRequestConfig) {
        return this.axios.post(
            "/application/{id}/measurement".replace(/{id}/, String(params["id"])), data, options
        );
    }

    private getMeasurement(params: {
        'id': string,
        'name': string
    }, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<Measurement>>(
            "/application/{id}/measurement/{name}".replace(/{id}/, String(params["id"])).replace(/{name}/, String(params["name"])), options
        );
    }

    private listMeasurements(params: {
        'id': string
    }, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<Measurement>[]>(
            "/application/{id}/measurements".replace(/{id}/, String(params["id"])), options
        );
    }

    private listMeasurementValues(params: {
        'id': string,
        'name': string,
        'sort': "created" | "value",
        'dir': "asc" | "desc",
        'page': number,
        'size': 10 | 20
    }, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<MeasurementValues>>(
            "/application/{id}/measurements/{name}".replace(/{id}/, String(params["id"])).replace(/{name}/, String(params["name"])),
            Object.assign(
                {},
                {
                    params: pick(params, "sort", "dir", "page", "size"),
                },
                options
            )
        );
    }

    private listApplications(params: Record<string, never>, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<Application>[]>(
            "/applications", options
        );
    }

    private login(params: Record<string, never>, data: WithoutReadonly<LoginRequest>, options?: AxiosRequestConfig) {
        return this.axios.post<WithoutWriteonly<LoginResponse>>(
            "/login", data, options
        );
    }

    private logout(params: {
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
            addApplication: this.addApplication.bind(this),
            getApplication: this.getApplication.bind(this),
            updateApplication: this.updateApplication.bind(this),
            deleteApplication: this.deleteApplication.bind(this),
            addConfiguration: this.addConfiguration.bind(this),
            addMeasurement: this.addMeasurement.bind(this),
            getMeasurement: this.getMeasurement.bind(this),
            listMeasurements: this.listMeasurements.bind(this),
            listMeasurementValues: this.listMeasurementValues.bind(this),
            listApplications: this.listApplications.bind(this),
            login: this.login.bind(this),
            logout: this.logout.bind(this)
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
