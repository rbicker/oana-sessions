<script>
  import { onDestroy, onMount } from "svelte";
  import { afterNavigate, goto } from "$app/navigation";
  import PocketBase from "pocketbase";

  let loading = true;
  let error = "";
  let sessions = [];
  let filteredSessions = [];
  let lastUpdated = "";
  let groupedSessions = [];
  let realtimeConnected = false;
  let categoriesByExternalId = {};
  let categoryOptions = [];
  let categoryVisibility = {};
  let filtersCollapsed = false;
  let weekStart = "";
  let currentWeekStart = "";
  let weekEnd = "";
  let weekRequestId = 0;
  let categoriesLoaded = false;
  const dayColors = [
    "#FFADAD",
    "#FFD6A5",
    "#9BF6FF",
    "#A0C4FF",
    "#BDB2FF",
    "#FFC6FF",
  ];

  const apiBase = import.meta.env.DEV
    ? "http://127.0.0.1:8090"
    : typeof window !== "undefined"
      ? window.location.origin
      : "";
  const pb = new PocketBase(apiBase);

  const collectionFilter = "source = 'coremanager' && source_category_id = 4";

  const formatDayWeekday = (day) =>
    new Date(`${day}T12:00:00Z`).toLocaleDateString(["de-CH"], {
      weekday: "long",
      timeZone: "UTC",
    });

  const formatDayDate = (day) =>
    new Date(`${day}T12:00:00Z`).toLocaleDateString(["de-CH"], {
      day: "2-digit",
      month: "2-digit",
      year: "2-digit",
      timeZone: "UTC",
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

  const normalizeSessionTitle = (value) =>
    (value || "").replace(/^1\. OANA Surf /, "");

  const summerPassCategories = new Set(["basic session", "advanced session"]);
  const isSummerPassCategory = (category) =>
    summerPassCategories.has(
      (category.title || "")
        .replace(/^(?:\d+\.\s*)?OANA Surf\s+/i, "")
        .trim()
        .toLowerCase(),
    );

  const toSessionViewModel = (record) => ({
    id: record.id,
    day: record.event_date || (record.start || "").split("T")[0],
    start: record.start || "",
    end: record.end || "",
    title: normalizeSessionTitle(record.title),
    categoryExternalId: record.category_external_id || "",
    categoryTitle:
      categoriesByExternalId[record.category_external_id || ""]?.title || "",
    categoryColor:
      categoriesByExternalId[record.category_external_id || ""]?.color ||
      "#94a3b8",
    participants: record.participants_count ?? 0,
    max_participants: record.max_participants ?? 0,
    isDeleted: record.is_deleted === true,
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

  const getWeekStart = (day) => {
    const weekday = new Date(`${day}T12:00:00Z`).getUTCDay();
    return addDaysToIsoDate(day, -(weekday === 0 ? 6 : weekday - 1));
  };

  const parseWeek = (value) => {
    if (!/^\d{4}-\d{2}-\d{2}$/.test(value || "")) return null;
    const date = new Date(`${value}T12:00:00Z`);
    if (Number.isNaN(date.getTime()) || date.toISOString().slice(0, 10) !== value) {
      return null;
    }
    return getWeekStart(value);
  };

  const navigateToWeek = (startDay, replaceState = false) => {
    const url = new URL(window.location.href);
    url.searchParams.set("week", startDay);
    void goto(url, { replaceState, noScroll: true, keepFocus: true });
  };

  const syncWeekFromUrl = (url) => {
    if (!categoriesLoaded) return;

    currentWeekStart = getWeekStart(getTodayZurich());
    const weekParameter = url.searchParams.get("week");
    const selectedWeek = parseWeek(weekParameter) || currentWeekStart;
    if (weekParameter !== selectedWeek) {
      navigateToWeek(selectedWeek, true);
    }
    if (weekStart !== selectedWeek) {
      void loadWeek(selectedWeek);
    }
  };

  afterNavigate(({ to }) => {
    if (to?.url) syncWeekFromUrl(to.url);
  });

  const loadWeek = async (startDay) => {
    const requestId = ++weekRequestId;
    weekStart = startDay;
    weekEnd = addDaysToIsoDate(startDay, 6);
    loading = true;
    error = "";

    try {
      const sessionRecords = await pb.collection("sessions").getFullList({
        sort: "start",
        filter: `${collectionFilter} && event_date >= '${startDay}' && event_date <= '${weekEnd}'`,
        fields:
          "id,event_date,start,end,title,category_external_id,participants_count,max_participants,participants_synced_at,last_synced_at,updated,is_deleted",
      });

      if (requestId !== weekRequestId) return;
      sessions = sessionRecords.map(toSessionViewModel);
      lastUpdated =
        sessions
          .map((session) => session.updated_hint)
          .filter(Boolean)
          .sort()
          .at(-1) || "";
    } catch (err) {
      if (requestId !== weekRequestId) return;
      error = err instanceof Error ? err.message : "Failed to load sessions";
    } finally {
      if (requestId === weekRequestId) loading = false;
    }
  };

  const setAllCategoriesVisibility = (visible) => {
    categoryVisibility = Object.fromEntries(
      categoryOptions.map((category) => [category.externalId, visible]),
    );
  };

  const selectSummerPass = () => {
    categoryVisibility = Object.fromEntries(
      categoryOptions.map((category) => [
        category.externalId,
        isSummerPassCategory(category),
      ]),
    );
  };

  const isSummerPassSelected = () =>
    categoryOptions.some(isSummerPassCategory) &&
    categoryOptions.every(
      (category) =>
        (categoryVisibility[category.externalId] !== false) ===
        isSummerPassCategory(category),
    );

  const toggleCategoryVisibility = (externalId) => {
    categoryVisibility = {
      ...categoryVisibility,
      [externalId]: !(categoryVisibility[externalId] ?? true),
    };
  };

  const areAllCategoriesVisible = () =>
    categoryOptions.length === 0 ||
    categoryOptions.every(
      (category) => categoryVisibility[category.externalId] !== false,
    );

  const syncFiltersCollapsedForViewport = () => {
    if (typeof window === "undefined") {
      return;
    }

    filtersCollapsed = window.innerWidth < 768;
  };

  const applyRealtimeUpdate = (event) => {
    const record = event?.record;
    if (!record) {
      return;
    }

    const todayZurich = getTodayZurich();
    const recordDay = record.event_date || (record.start || "").split("T")[0];
    if (
      recordDay !== todayZurich ||
      todayZurich < weekStart ||
      todayZurich > weekEnd
    ) {
      return;
    }

    if (event.action === "delete") {
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
    filteredSessions = sessions.filter(
      (session) => categoryVisibility[session.categoryExternalId] !== false,
    );
  }

  $: {
    const grouped = {};
    for (const session of filteredSessions) {
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
    syncFiltersCollapsedForViewport();
    window.addEventListener("resize", syncFiltersCollapsedForViewport);

    try {
      const todayZurich = getTodayZurich();
      const categoryRecords = await pb.collection("categories").getFullList({
        sort: "position,title",
        filter: `is_deleted = false && ${collectionFilter}`,
        fields: "external_id,title,color,position",
      });

      categoriesByExternalId = Object.fromEntries(
        categoryRecords.map((category) => [
          category.external_id,
          { title: category.title, color: category.color },
        ]),
      );
      categoryOptions = categoryRecords.map((category) => ({
        externalId: category.external_id,
        title: category.title,
        color: category.color,
      }));
      setAllCategoriesVisibility(true);

      categoriesLoaded = true;
      syncWeekFromUrl(new URL(window.location.href));

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
      loading = false;
    }
  });

  onDestroy(() => {
    pb.collection("sessions").unsubscribe("*");
    if (typeof window !== "undefined") {
      window.removeEventListener("resize", syncFiltersCollapsedForViewport);
    }
  });
</script>

<svelte:head>
  <title>Oana Sessions</title>
</svelte:head>

<main class="sessions-page" data-theme="oana">
  <!--
  <nav class="navbar sessions-nav" aria-label="Session navigation">
    <div class="navbar-start sessions-brand">
      <h1>Oana Sessions</h1>
    </div>

    <div class="navbar-end sessions-nav-tools">
      <a class="sessions-nav-link" href="/">Sessions</a>
    </div>
  </nav>
  -->
  {#if weekStart}
    <nav class="week-navigation" aria-label="Wochennavigation">
      <button
        type="button"
        class="btn btn-outline"
        on:click={() => navigateToWeek(addDaysToIsoDate(weekStart, -7))}
        aria-label="Vorherige Woche"
      >
        ← <span>Vorherige Woche</span>
      </button>
      <div class="week-navigation-current" aria-live="polite">
        <strong>{formatDayDate(weekStart)} – {formatDayDate(weekEnd)}</strong>
        <button
          type="button"
          class="week-today-button"
          disabled={weekStart === currentWeekStart}
          on:click={() => navigateToWeek(getWeekStart(getTodayZurich()))}
        >
          Diese Woche
        </button>
      </div>
      <button
        type="button"
        class="btn btn-outline"
        on:click={() => navigateToWeek(addDaysToIsoDate(weekStart, 7))}
        aria-label="Nächste Woche"
      >
        <span>Nächste Woche</span> →
      </button>
    </nav>
  {/if}
  {#if categoryOptions.length > 0}
    <section class="sessions-filters" aria-label="Session category filters">
      <div class="sessions-filters-header">
        <button
          class="sessions-filters-button"
          type="button"
          aria-expanded={!filtersCollapsed}
          on:click={() => (filtersCollapsed = !filtersCollapsed)}
        >
          <span>Kategorien</span>
          <span class="sessions-filters-chevron" class:is-collapsed={filtersCollapsed}
            >▾</span
          >
        </button>
        <div class="sessions-filter-shortcuts">
          <label class="sessions-filter-toggle">
            <input
              type="checkbox"
              checked={isSummerPassSelected()}
              on:change={(event) =>
                event.currentTarget.checked
                  ? selectSummerPass()
                  : setAllCategoriesVisibility(true)}
            />
            <span>Summer Pass</span>
          </label>
          <label class="sessions-filter-toggle">
            <input
              type="checkbox"
              checked={areAllCategoriesVisible()}
              on:change={(event) =>
                setAllCategoriesVisibility(event.currentTarget.checked)}
            />
            <span>Alle anzeigen</span>
          </label>
        </div>
      </div>

      {#if !filtersCollapsed}
        <div class="sessions-filter-list">
          {#each categoryOptions as category}
            <label class="sessions-filter-item">
              <input
                type="checkbox"
                checked={categoryVisibility[category.externalId] !== false}
                on:change={() => toggleCategoryVisibility(category.externalId)}
              />
              <span
                class="session-category-dot"
                style={`background: ${category.color};`}
                aria-hidden="true"
              ></span>
              <span>{category.title}</span>
            </label>
          {/each}
        </div>
      {/if}
    </section>
  {/if}
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
  {:else if filteredSessions.length === 0}
    <div class="alert shadow-sm sessions-alert">
      <span>Keine Sessions für die ausgewählten Kategorien.</span>
    </div>
  {:else}
    <div class="day-grid">
      {#each groupedSessions as group, i}
        <section class="session-day">
          <article
            class="card sessions-day-card"
            style={`--day-color: ${dayColors[i % dayColors.length]};`}
          >
            <div class="card-body day-card-body">
              <h2 class="card-title day-card-title">
                <span class="day-card-weekday"
                  >{formatDayWeekday(group.day)}</span
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
                      <tr class:session-deleted={session.isDeleted}>
                        <td class="session-time"
                          >{formatTime(session.start)} - {formatTime(
                            session.end,
                          )}</td
                        >
                        <td>
                          <div
                            class="session-title"
                            title={session.categoryTitle || session.title}
                          >
                            <span
                              class="session-category-dot"
                              style={`background: ${session.categoryColor};`}
                              aria-hidden="true"
                            ></span>
                            <span class="session-title-text">{session.title}</span>
                            {#if session.isDeleted}<span class="sr-only"> (gelöscht)</span>{/if}
                          </div>
                        </td>
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

<style>
  .sessions-filters {
    display: grid;
    gap: 0.85rem;
    margin: 1rem 0 1.25rem;
    padding: 1rem 1.1rem;
    border: 1px solid rgb(15 23 42 / 0.08);
    border-radius: 1rem;
    background: rgb(255 255 255 / 0.72);
    backdrop-filter: blur(10px);
  }

  .week-navigation {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 0.75rem;
    margin: 0 0 1rem;
  }

  .week-navigation-current {
    display: flex;
    flex-direction: column;
    align-items: center;
    text-align: center;
    gap: 0.2rem;
    color: var(--color-neutral);
  }

  .week-today-button {
    border: 0;
    padding: 0;
    background: transparent;
    color: var(--color-primary);
    text-decoration: underline;
    cursor: pointer;
  }

  .week-today-button:disabled {
    opacity: 0.55;
    cursor: default;
    text-decoration: none;
  }

  @media (max-width: 640px) {
    .week-navigation .btn {
      min-height: 2.5rem;
      padding: 0 0.75rem;
    }

    .week-navigation .btn span {
      display: none;
    }
  }

  .sessions-filters-header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .sessions-filter-shortcuts {
    display: flex;
    align-items: center;
    gap: 1rem;
    flex-wrap: wrap;
  }

  .sessions-filters-button {
    display: inline-flex;
    align-items: center;
    gap: 0.55rem;
    padding: 0;
    margin: 0;
    border: 0;
    background: transparent;
    color: inherit;
    font: inherit;
    font-size: 0.95rem;
    font-weight: 700;
    text-transform: uppercase;
    letter-spacing: 0.08em;
    cursor: pointer;
  }

  .sessions-filters-chevron {
    transition: transform 160ms ease;
  }

  .sessions-filters-chevron.is-collapsed {
    transform: rotate(-90deg);
  }

  .sessions-filter-list {
    display: grid;
    gap: 0.65rem;
  }

  .sessions-filter-item,
  .sessions-filter-toggle {
    display: inline-flex;
    align-items: center;
    gap: 0.55rem;
    font-size: 0.95rem;
    cursor: pointer;
  }

  .sessions-filter-item input,
  .sessions-filter-toggle input {
    margin: 0;
  }

  .sessions-filter-item {
    width: 100%;
  }

  .session-title {
    display: inline-flex;
    align-items: center;
    gap: 0.55rem;
  }

  .session-deleted {
    opacity: 0.6;
  }

  .session-deleted .session-time,
  .session-deleted .session-title-text,
  .session-deleted .participants-cell {
    text-decoration: line-through;
  }

  .session-category-dot {
    width: 0.7rem;
    height: 0.7rem;
    border-radius: 999px;
    flex: 0 0 auto;
    box-shadow: 0 0 0 1px rgb(15 23 42 / 0.12);
  }

  @media (min-width: 768px) {
    .sessions-filter-list {
      grid-template-columns: repeat(auto-fit, minmax(15rem, 1fr));
      gap: 0.65rem 1rem;
    }

    .sessions-filter-item {
      width: auto;
    }
  }
</style>
