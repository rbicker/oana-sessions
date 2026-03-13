<script>
  import { onDestroy, onMount } from "svelte";
  import PocketBase from "pocketbase";

  const colors = [
    "#FFADAD",
    "#FFD6A5",
    "#9BF6FF",
    "#A0C4FF",
    "#BDB2FF",
    "#FFC6FF",
  ];

  let loading = true;
  let error = "";
  let sessions = [];
  let lastUpdated = "";
  let groupedSessions = [];
  let realtimeConnected = false;

  const apiBase = import.meta.env.DEV
    ? "http://127.0.0.1:8090"
    : typeof window !== "undefined"
      ? window.location.origin
      : "";
  const pb = new PocketBase(apiBase);

  const collectionFilter =
    "is_deleted = false && source = 'coremanager' && source_category_id = 4";

  const formatDayWeekday = (day) =>
    new Date(day).toLocaleDateString(["de-CH"], { weekday: "long" });

  const formatDayDate = (day) =>
    new Date(day).toLocaleDateString(["de-CH"], {
      day: "2-digit",
      month: "2-digit",
      year: "2-digit",
    });

  const toDate = (value) => {
    const parsed = new Date(value);
    if (!Number.isNaN(parsed.getTime())) {
      return parsed;
    }

    if (typeof value === "string" && value.includes(" ")) {
      const withT = value.replace(" ", "T");
      const withZ = withT.endsWith("Z") ? withT : `${withT}Z`;
      const fallback = new Date(withZ);
      if (!Number.isNaN(fallback.getTime())) {
        return fallback;
      }
    }

    return null;
  };

  const formatTime = (value) => {
    const parsed = toDate(value);
    if (!parsed) {
      return "-";
    }

    return parsed.toLocaleTimeString(["de-CH"], {
      hour: "2-digit",
      minute: "2-digit",
    });
  };

  const formatDateTime = (value) => {
    const parsed = toDate(value);
    if (!parsed) {
      return "-";
    }
    return parsed.toLocaleString(["de-CH"]);
  };

  const toSessionViewModel = (record) => ({
    id: record.id,
    day: record.event_date || (record.start || "").split("T")[0],
    start: record.start || "",
    end: record.end || "",
    title: record.title || "",
    participants: record.participants_count ?? 0,
    max_participants: record.max_participants ?? 0,
    updated_hint:
      record.participants_synced_at ||
      record.last_synced_at ||
      record.updated ||
      "",
  });

  const getTodayZurich = () =>
    new Intl.DateTimeFormat("en-CA", {
      timeZone: "Europe/Zurich",
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    }).format(new Date());

  const addDaysToIsoDate = (isoDate, daysToAdd) => {
    const [year, month, day] = isoDate.split("-").map(Number);
    const date = new Date(Date.UTC(year, month - 1, day));
    date.setUTCDate(date.getUTCDate() + daysToAdd);
    return date.toISOString().slice(0, 10);
  };

  const applyRealtimeUpdate = (event) => {
    const record = event?.record;
    if (!record) {
      return;
    }

    const todayZurich = getTodayZurich();
    const recordDay = record.event_date || (record.start || "").split("T")[0];
    if (recordDay !== todayZurich) {
      return;
    }

    if (event.action === "delete" || record.is_deleted) {
      sessions = sessions.filter((session) => session.id !== record.id);
      return;
    }

    const mapped = toSessionViewModel(record);
    const index = sessions.findIndex((session) => session.id === mapped.id);
    if (index === -1) {
      sessions = [...sessions, mapped].sort((a, b) =>
        a.start.localeCompare(b.start),
      );
    } else {
      sessions = sessions.map((session) =>
        session.id === mapped.id ? mapped : session,
      );
    }

    lastUpdated = mapped.updated_hint || new Date().toISOString();
  };

  $: {
    const grouped = {};
    for (const session of sessions) {
      const day = session.day || (session.start || "").split("T")[0];
      if (!day) {
        continue;
      }
      if (!grouped[day]) {
        grouped[day] = [];
      }
      grouped[day].push(session);
    }

    const sortedDays = Object.keys(grouped).sort((a, b) => a.localeCompare(b));
    groupedSessions = sortedDays.map((day) => ({
      day,
      sessions: grouped[day],
    }));
  }

  onMount(async () => {
    try {
      const todayZurich = getTodayZurich();
      const endDay = addDaysToIsoDate(todayZurich, 6);
      const dateRangeFilter = `event_date >= '${todayZurich}' && event_date <= '${endDay}'`;

      const records = await pb.collection("sessions").getFullList({
        sort: "start",
        filter: `${collectionFilter} && ${dateRangeFilter}`,
        fields:
          "id,event_date,start,end,title,participants_count,max_participants,participants_synced_at,last_synced_at,updated,is_deleted",
      });

      sessions = records.map(toSessionViewModel);
      lastUpdated =
        sessions
          .map((session) => session.updated_hint)
          .filter(Boolean)
          .sort()
          .at(-1) || "";

      try {
        await pb.collection("sessions").subscribe(
          "*",
          (event) => {
            applyRealtimeUpdate(event);
          },
          {
            filter: `${collectionFilter} && event_date = '${todayZurich}'`,
          },
        );
        realtimeConnected = true;
      } catch (realtimeErr) {
        console.error("Realtime subscription failed", realtimeErr);
      }
    } catch (err) {
      error = err instanceof Error ? err.message : "Failed to load sessions";
    } finally {
      loading = false;
    }
  });

  onDestroy(() => {
    pb.collection("sessions").unsubscribe("*");
  });
</script>

<svelte:head>
  <title>Oana Sessions</title>
</svelte:head>

<main class="sessions-page" data-theme="oana">
  <nav class="navbar sessions-nav" aria-label="Session navigation">
    <div class="navbar-start sessions-brand">
      <h1>Oana Sessions</h1>
    </div>

    <div class="navbar-end sessions-nav-tools">
      <a class="sessions-nav-link" href="/">Sessions</a>
    </div>
  </nav>
  {#if loading}
    <div class="alert shadow-sm sessions-alert">
      <span>Sessions werden aktuell geladen.</span>
    </div>
  {:else if error}
    <div class="alert alert-error shadow-sm sessions-alert">
      <span>{error}</span>
    </div>
  {:else if sessions.length === 0}
    <div class="alert shadow-sm sessions-alert">
      <span>Keine Sessions verfügbar.</span>
    </div>
  {:else}
    <div class="day-grid">
      {#each groupedSessions as group, i}
        <section class="session-day">
          <article class="card sessions-day-card"
            style={`--day-color: ${colors[i % colors.length]};`}
          >
            <div class="card-body day-card-body">
            <h2 class="card-title day-card-title">
              <span class="day-card-weekday">{formatDayWeekday(group.day)}</span
              >
              <span class="day-card-date">{formatDayDate(group.day)}</span>
            </h2>

            <div class="overflow-x-auto day-card-table-wrap">
              <table class="table table-sm sessions-table">
                <thead>
                  <tr>
                    <th>Zeitslot</th>
                    <th>Session</th>
                    <th>Teilnehmer</th>
                  </tr>
                </thead>
                <tbody>
                  {#each group.sessions as session}
                    <tr>
                      <td
                        >{formatTime(session.start)} - {formatTime(
                          session.end,
                        )}</td
                      >
                      <td>{session.title}</td>
                      <td class="participants-cell"
                        >{session.participants}/{session.max_participants}</td
                      >
                    </tr>
                  {/each}
                </tbody>
              </table>
            </div>
          </div>
          </article>
        </section>
      {/each}
    </div>

    <footer class="sessions-footer">
      Last updated: {lastUpdated ? formatDateTime(lastUpdated) : "-"}
    </footer>
  {/if}
</main>
