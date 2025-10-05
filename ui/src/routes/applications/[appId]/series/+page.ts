import type { Series } from "$lib/client/definitions";
import { ListSeries } from "$lib/service";
import type { PageLoad } from "./$types";



const loadSeries = async function(appId: string): Promise<Series[]> {
	return await ListSeries(appId)
}

export const load: PageLoad = async ({ params }) => {
	let series: Series[] = await loadSeries(params.appId)

	return {
		series: series.map((n) => {
			return {
				value: n.name,
				name: n.name
			}
		}),
		appId: params.appId
	};
};