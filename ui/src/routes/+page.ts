import type { PageLoad } from './$types';
import { ListApplications } from '../service';

export const load: PageLoad = async ({ parent }) => {
	await parent()
	return { apps: await ListApplications() };
};