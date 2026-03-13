import { n as onDestroy } from "../../chunks/index-server.js";
import { i as head } from "../../chunks/server.js";
import PocketBase from "pocketbase";
//#region src/routes/+page.svelte
function _page($$renderer, $$props) {
	$$renderer.component(($$renderer) => {
		let sessions = [];
		const pb = new PocketBase(typeof window !== "undefined" ? window.location.origin : "");
		onDestroy(() => {
			pb.collection("sessions").unsubscribe("*");
		});
		$: {
			const grouped = {};
			for (const session of sessions) {
				const day = session.day || (session.start || "").split("T")[0];
				if (!day) continue;
				if (!grouped[day]) grouped[day] = [];
				grouped[day].push(session);
			}
			Object.keys(grouped).sort((a, b) => a.localeCompare(b)).map((day) => ({
				day,
				sessions: grouped[day]
			}));
		}
		head("1uha8ag", $$renderer, ($$renderer) => {
			$$renderer.title(($$renderer) => {
				$$renderer.push(`<title>Surf Sessions</title>`);
			});
		});
		$$renderer.push(`<main class="sessions-page">`);
		$$renderer.push("<!--[0-->");
		$$renderer.push(`<div class="alert"><span>Sessions are currently loading. Please check back in a few minutes.</span></div>`);
		$$renderer.push(`<!--]--></main>`);
	});
}
//#endregion
export { _page as default };
