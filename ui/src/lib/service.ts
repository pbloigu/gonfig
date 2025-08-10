import { AxiosError, AxiosHeaders, type AxiosResponse } from "axios";
import Api from "./client/api"
import { type Application, type LoginResponse, type Measurement, type MeasurementValues, type StatusChangeTrigger, type WithoutWriteonly } from './client/definitions'
import { goto } from "$app/navigation";
import { get } from "svelte/store";
import { error } from "@sveltejs/kit";



var baseUrl: string | undefined = undefined


const getApi = function (): Api {
    var a = new Api({
        baseURL: baseUrl
    });
    a.axios.interceptors.request.use((r) => {
        r.headers.set("Authorization", "Bearer " + localStorage.getItem("apiKey"))
        return r
    })

    return a
}

const defaultErrorHandler = async function (error: AxiosError) {
        await goto("/loginform")
        return Promise.reject(error)
}

const notFoundErrorHandler = async function (error: AxiosError) {
    if (error.response?.status == 404) {
        return Promise.resolve(null)
    } else {
        return await defaultErrorHandler(error)
    }
}

export const InitApi = function (baseUrlToSet: string) {
    if (baseUrl == undefined) {
        baseUrl = baseUrlToSet
    }
}

export const Logout = async function () {
    const response: AxiosResponse<any, any> = await getApi().Default.logout(
        { Authorization: localStorage.getItem("apiKey") || "" }, {})
    return Promise.resolve(response.data).then((l: any) => {
        goto("/loginform")
    })
}

export const Login = async function name(username: string, password: string) {
    const response: AxiosResponse<any, any> = await getApi().Default.login({}, { username: username, password: password })
        .catch((error: AxiosError) => defaultErrorHandler(error))
    Promise.resolve(response.data).then((l: LoginResponse) => {
        localStorage.setItem("apiKey", l.token)
        goto("/")
    })
}

export const ListApplications = async function (): Promise<Application[]> {
    const response: AxiosResponse<WithoutWriteonly<Application>[]> = await getApi().Default.listApplications({}, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data);
}

export const AddApplication = async function (app: Application): Promise<Application> {
    const response = await getApi().Default.addApplication({}, app, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const AddMeasurement = async function name(name: string, appId: string): Promise<Measurement> {
    const response = await getApi().Default.addMeasurement({ id: appId }, { name: name }, {})
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

export const ListMeasurements = async function (appId: string): Promise<Measurement[]> {
    const response = await getApi().Default.listMeasurements({ id: appId }, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const ListMeasurementValues = async function name(name: string, appId: string, page: number): Promise<MeasurementValues> {
    const response = await getApi().Default.listMeasurementValues({
            id: appId, 
            name: name,
            dir: "desc",
            page: page,
            size:10,
            sort:"data"
        }, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)

}

export const GetStatusChangeTrigger = async function(appId: string): Promise<StatusChangeTrigger | null> {
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

export const AddStatusChangeTrigger = async function(appId: string, trigger: StatusChangeTrigger): Promise<StatusChangeTrigger> {
    const response = await getApi().Default.addStatusChangeTrigger({id: appId}, trigger, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const UpdateStatusChangeTrigger = async function(appId: string, trigger: StatusChangeTrigger) {
    const response = await getApi().Default.updateStatusChangeTrigger({id: appId}, trigger, {})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data)
}

export const DeleteStatusChangeTrigger = async function (appId: string) {
    const response = await getApi().Default.deleteStatusChangeTrigger({id: appId})
        .catch((error: AxiosError) => defaultErrorHandler(error))
    return Promise.resolve(response.data) 
}