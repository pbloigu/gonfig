
import { ListApplications } from '$lib/service';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ parent }) => {
	await parent()
	return { apps: await ListApplications() };
};