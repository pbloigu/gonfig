export const ssr = false;
export const prerender = true;
import { InitApi } from '../service';
import type { PageLoad } from './$types';
import { ListApplications } from '../service';
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