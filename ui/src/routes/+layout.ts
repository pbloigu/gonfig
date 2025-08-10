export const ssr = false;
export const prerender = true;
import { InitApi, ListApplications } from '$lib/service';
import type { PageLoad } from './$types';
const { MODE } = import.meta.env;


export const load: PageLoad = async () => {
	InitApi(resolveBaseUri())
};

function resolveBaseUri(): string {
	if (MODE == "development") {
		return "http://localhost:8080"
	} else {
		return window.location.protocol + "//" + window.location.host
	}
}