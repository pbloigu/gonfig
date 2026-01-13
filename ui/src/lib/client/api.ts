/* tslint:disable */
/* eslint-disable */

import axios, { type AxiosInstance, type AxiosRequestConfig } from "axios";
import type { Application, Configuration, Series, SeriesValues, StatusChangeTrigger, CronTrigger, CronValidationRequest, LoginRequest, LoginResponse, WithoutReadonly, WithoutWriteonly } from "./definitions";

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

    private isOnline(params: {
        'id': string
    }, options?: AxiosRequestConfig) {
        return this.axios.get(
            "/application/{id}/online".replace(/{id}/, String(params["id"])), options
        );
    }

    private listSeries(params: {
        'id': string
    }, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<Series>[]>(
            "/application/{id}/series".replace(/{id}/, String(params["id"])), options
        );
    }

    private addSeries(params: {
        'id': string
    }, data: WithoutReadonly<Series>, options?: AxiosRequestConfig) {
        return this.axios.post(
            "/application/{id}/series".replace(/{id}/, String(params["id"])), data, options
        );
    }

    private getSeries(params: {
        'id': string,
        'name': string
    }, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<Series>>(
            "/application/{id}/series/{name}".replace(/{id}/, String(params["id"])).replace(/{name}/, String(params["name"])), options
        );
    }

    private listSeriesValues(params: {
        'id': string,
        'name': string,
        'sort': "created" | "data" | " recorded",
        'dir': "asc" | "desc",
        'page': number,
        'size': number
    }, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<SeriesValues>>(
            "/application/{id}/series/{name}/values".replace(/{id}/, String(params["id"])).replace(/{name}/, String(params["name"])),
            Object.assign(
                {},
                {
                    params: pick(params, "sort", "dir", "page", "size"),
                },
                options
            )
        );
    }

    private getStatusChangeTrigger(params: {
        'id': string
    }, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<StatusChangeTrigger>>(
            "/application/{id}/trigger/status".replace(/{id}/, String(params["id"])), options
        );
    }

    private addStatusChangeTrigger(params: {
        'id': string
    }, data: WithoutReadonly<StatusChangeTrigger>, options?: AxiosRequestConfig) {
        return this.axios.post<WithoutWriteonly<StatusChangeTrigger>>(
            "/application/{id}/trigger/status".replace(/{id}/, String(params["id"])), data, options
        );
    }

    private updateStatusChangeTrigger(params: {
        'id': string
    }, data: WithoutReadonly<StatusChangeTrigger>, options?: AxiosRequestConfig) {
        return this.axios.put<WithoutWriteonly<StatusChangeTrigger>>(
            "/application/{id}/trigger/status".replace(/{id}/, String(params["id"])), data, options
        );
    }

    private deleteStatusChangeTrigger(params: {
        'id': string
    }, options?: AxiosRequestConfig) {
        return this.axios.delete(
            "/application/{id}/trigger/status".replace(/{id}/, String(params["id"])), options
        );
    }

    private listApplications(params: Record<string, never>, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<Application>[]>(
            "/applications", options
        );
    }

    private addCronTrigger(params: Record<string, never>, data: WithoutReadonly<CronTrigger>, options?: AxiosRequestConfig) {
        return this.axios.post<WithoutWriteonly<CronTrigger>>(
            "/cron", data, options
        );
    }

    private isValid(params: Record<string, never>, data: WithoutReadonly<CronValidationRequest>, options?: AxiosRequestConfig) {
        return this.axios.post(
            "/cron/expression", data, options
        );
    }

    private updateCronTrigger(params: {
        'id': number
    }, data: WithoutReadonly<CronTrigger>, options?: AxiosRequestConfig) {
        return this.axios.put<WithoutWriteonly<CronTrigger>>(
            "/cron/{id}".replace(/{id}/, String(params["id"])), data, options
        );
    }

    private deleteCronTrigger(params: {
        'id': number
    }, options?: AxiosRequestConfig) {
        return this.axios.delete(
            "/cron/{id}".replace(/{id}/, String(params["id"])), options
        );
    }

    private listCronTriggers(params: Record<string, never>, options?: AxiosRequestConfig) {
        return this.axios.get<WithoutWriteonly<CronTrigger>[]>(
            "/crons", options
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
            isOnline: this.isOnline.bind(this),
            listSeries: this.listSeries.bind(this),
            addSeries: this.addSeries.bind(this),
            getSeries: this.getSeries.bind(this),
            listSeriesValues: this.listSeriesValues.bind(this),
            getStatusChangeTrigger: this.getStatusChangeTrigger.bind(this),
            addStatusChangeTrigger: this.addStatusChangeTrigger.bind(this),
            updateStatusChangeTrigger: this.updateStatusChangeTrigger.bind(this),
            deleteStatusChangeTrigger: this.deleteStatusChangeTrigger.bind(this),
            listApplications: this.listApplications.bind(this),
            addCronTrigger: this.addCronTrigger.bind(this),
            isValid: this.isValid.bind(this),
            updateCronTrigger: this.updateCronTrigger.bind(this),
            deleteCronTrigger: this.deleteCronTrigger.bind(this),
            listCronTriggers: this.listCronTriggers.bind(this),
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
