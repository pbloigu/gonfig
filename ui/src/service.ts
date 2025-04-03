import { AxiosError, AxiosHeaders, type AxiosResponse } from "axios";
import Api from "./client/api"
import { type Application, type LoginResponse, type Measurement, type WithoutWriteonly } from './client/definitions'
import { goto } from "$app/navigation";



var baseUrl: string | undefined = undefined


const getApi = function(): Api {
    var a = new Api({
        baseURL: baseUrl
    });
    a.axios.interceptors.request.use((r) => {
        r.headers.set("Authorization", "Bearer "+localStorage.getItem("apiKey"))
        return r
    })

    return a
}

const errorHandler = async function(error: AxiosError) {
    await goto("/loginform")
    return Promise.reject(error)
}

export const InitApi = function (baseUrlToSet: string) {
    if (baseUrl == undefined) {
        baseUrl = baseUrlToSet
    }
}

export const Logout = async function (){
    const response: AxiosResponse<any, any> = await getApi().Default.logout_get(
        {Authorization: localStorage.getItem("apiKey") || ""}, {})
    return Promise.resolve(response.data).then((l: any) => {
        goto("/loginform")
    })
}

export const Login = async function name(username: string, password: string){
    const response: AxiosResponse<any, any> = await getApi().Default.login_post({}, {username: username, password: password})
    .catch((error: AxiosError) => errorHandler(error))
    Promise.resolve(response.data).then((l: LoginResponse) => {
        localStorage.setItem("apiKey", l.token)
        goto("/")
    })
}

export const ListApplications = async function (): Promise<Application[]> {
    const response: AxiosResponse<WithoutWriteonly<Application>[]> = await getApi().Default.applications_get({}, {})
    .catch((error: AxiosError) => errorHandler(error))
    return Promise.resolve(response.data);
}

export const AddApplication = async function (app: Application): Promise<Application> {
    const response = await getApi().Default.application_post({}, app, {})
    .catch((error: AxiosError) => errorHandler(error))
    return Promise.resolve(response.data)
}

export const UpdateConfiguration = async function (appId: string, configuration: string) {
    const response = await getApi().Default.configuration_post({ id: appId }, { data: configuration }, {})
    .catch((error: AxiosError) => errorHandler(error))
    return Promise.resolve(response.data)
}

export const DeleteApplication = async function (appId: string) {
    const response = await getApi().Default.id_delete({ id: appId }, {})
    .catch((error: AxiosError) => errorHandler(error))
    return Promise.resolve(response.data)
}

export const ListMeasurements = async function(appId: string): Promise<Measurement[]> {
    const response = await getApi().Default.measurements_get({id: appId},{})
    .catch((error: AxiosError) => errorHandler(error))
    return Promise.resolve(response.data)
}