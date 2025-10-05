import { AxiosError, AxiosHeaders, type AxiosResponse } from "axios";
import Api from "./client/api"
import { type Application, type LoginResponse, type Series, type SeriesValues, type StatusChangeTrigger, type WithoutWriteonly } from './client/definitions'
import { goto } from "$app/navigation";
const { MODE, VITE_API_URL } = import.meta.env;


export interface Error {
    code?: number
    message?: string
}


var baseUrl: string | undefined = undefined


const resolveBaseUri = function(): string {
    if (MODE == "development") {
		return VITE_API_URL
	} else {
		return window.location.protocol + "//" + window.location.host
	}
}

const getApi = function (): Api {
    var a = new Api({
        baseURL: resolveBaseUri()
    });
    a.axios.interceptors.request.use((r) => {
        r.headers.set("Authorization", "Bearer " + localStorage.getItem("apiKey"))
        return r
    })

    return a
}

const defaultErrorHandler = async function (error: AxiosError) {
    if(error.response?.status == 401) {
        await goto("/")
    } else {
        await goto("/error", {
            state: {
                code: error.response?.status,
                message: error.cause
            }
        })
    }
    return Promise.reject(error)
}

const notFoundErrorHandler = async function (error: AxiosError) {
    if (error.response?.status == 404) {
        return Promise.resolve(null)
    } else {
        return await defaultErrorHandler(error)
    }
}


export const Logout = async function () {
    const response: AxiosResponse<any, any> = await getApi().Default.logout(
        { Authorization: localStorage.getItem("apiKey") || "" }, {})
    return Promise.resolve(response.data).then((l: any) => {
        goto("/")
    })
}

export const Login = async function name(username: string, password: string) {
    const response: AxiosResponse<any, any> = await getApi().Default.login({}, { username: username, password: password })
        .catch((error: AxiosError) => defaultErrorHandler(error))
    Promise.resolve(response.data).then((l: LoginResponse) => {
        localStorage.setItem("apiKey", l.token)
        goto("/applications")
    })
}

export const ListApplications = async function (): Promise<Application[]> {
    const response: AxiosResponse<WithoutWriteonly<Application>[]> = await getApi().Default.listApplications({}, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data);
}

export const GetApplication = async function (appId: string): Promise<Application> {
    const response: AxiosResponse<WithoutWriteonly<Application>> = await getApi().Default.getApplication({ id: appId }, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const AddApplication = async function (app: Application): Promise<Application> {
    const response = await getApi().Default.addApplication({}, app, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const AddSeries = async function name(name: string, appId: string): Promise<Series> {
    const response = await getApi().Default.addSeries({ id: appId }, { name: name }, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const UpdateConfiguration = async function (appId: string, configuration: string) {
    const response = await getApi().Default.addConfiguration({ id: appId }, { data: configuration }, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const DeleteApplication = async function (appId: string) {
    const response = await getApi().Default.deleteApplication({ id: appId }, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const ListSeries = async function (appId: string): Promise<Series[]> {
    const response = await getApi().Default.listSeries({ id: appId }, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const ListSeriesValues = async function name(name: string, appId: string, page: number): Promise<SeriesValues> {
    const response = await getApi().Default.listSeriesValues({
        id: appId,
        name: name,
        dir: "desc",
        page: page,
        size: 10,
        sort: "data"
    }, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)

}

export const GetStatusChangeTrigger = async function (appId: string): Promise<StatusChangeTrigger | null> {
    const response = await getApi().Default.getStatusChangeTrigger({ id: appId }, {})
        .catch((error: AxiosError) => notFoundErrorHandler(error))
    return Promise.resolve(response).then((r: AxiosResponse | null) => {
        if (r == null) {
            return null
        } else {
            return r.data
        }
    })
}

export const AddStatusChangeTrigger = async function (appId: string, trigger: StatusChangeTrigger): Promise<StatusChangeTrigger> {
    const response = await getApi().Default.addStatusChangeTrigger({ id: appId }, trigger, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const UpdateStatusChangeTrigger = async function (appId: string, trigger: StatusChangeTrigger) {
    const response = await getApi().Default.updateStatusChangeTrigger({ id: appId }, trigger, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const DeleteStatusChangeTrigger = async function (appId: string) {
    const response = await getApi().Default.deleteStatusChangeTrigger({ id: appId })
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}