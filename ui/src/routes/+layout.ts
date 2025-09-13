export const ssr = false;
export const prerender = true;
import { InitApi } from '$lib/service';
import type { PageLoad } from './$types';
const { MODE, VITE_API_URL } = import.meta.env;


export const load: PageLoad = async () => {
	InitApi(resolveBaseUri())
};

function resolveBaseUri(): string {
	if (MODE == "development") {
		return VITE_API_URL
	} else {
		return window.location.protocol + "//" + window.location.host
	}
}