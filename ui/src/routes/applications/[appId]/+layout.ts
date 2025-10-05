import type { Application } from "$lib/client/definitions";
import { GetApplication } from "$lib/service";
import type { PageLoad } from "../$types";



export const load: PageLoad = async ({ params }) => {

    let app: Application = await GetApplication(params.appId)
    

    return {
        app: app
    };
};