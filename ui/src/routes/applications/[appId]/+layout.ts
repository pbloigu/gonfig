import type { components } from "$lib/client/api";
import { GetApplication } from "$lib/service";
import type { PageLoad } from "../$types";



export const load: PageLoad = async ({ params }) => {

    let app: components["schemas"]["Application"] = await GetApplication(params.appId)
    return {
        app: app
    };
};