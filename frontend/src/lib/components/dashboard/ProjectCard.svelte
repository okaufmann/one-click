<script lang="ts">
  import { goto } from "$app/navigation";
  import type { ProjectsResponse, RolloutsResponse } from "$lib/pocketbase/generated-types";
  import type { Pexpand, Rexpand } from "$lib/stores/data";
  import { rollouts } from "$lib/stores/data";
  import selectedProjectId from "$lib/stores/project";
  import { formatDateTime, timeAgo } from "$lib/utils/date.utils";
  import { recordLogoUrl } from "$lib/utils/blueprint.utils";
  import { Badge, Button, Indicator, Tooltip } from "flowbite-svelte";
  import { ArrowRight, Cog, ExternalLink, Tag } from "lucide-svelte";
  import type { RolloutStatusResponse } from "$lib/types/status";
  import { getRolloutStatus } from "$lib/utils/rollouts";
  import { onMount } from "svelte";
  import { navigating } from "$app/stores";
  export let project: ProjectsResponse<Pexpand>;

  let tags: Set<string> = new Set();
  if (project.tags) {
    tags = new Set(project.tags.split(","));
  }

  let current_rollout_status: RolloutStatusResponse | undefined;
  let rollout_status_color:
    | "gray"
    | "red"
    | "yellow"
    | "green"
    | "indigo"
    | "purple"
    | "blue"
    | "dark"
    | "orange"
    | "none"
    | "teal"
    | undefined;

  const determineRolloutColor = (status: string) => {
    switch (status) {
      case "Pending":
        return "yellow";
      case "Not Ready":
        return "yellow";
      case "Error":
        return "red";
      case "OK":
        return "green";
      default:
        return "gray";
    }
  };

  const updateCurrentRollout = () => {
    // find the rollout with no endDate of the selected project
    let currentRollout = $rollouts.find((r) => r.project === project.id && !r.endDate);

    getRolloutStatus($selectedProjectId, currentRollout?.id ?? "")
      .then((response) => {
        current_rollout_status = response;
        rollout_status_color = determineRolloutColor(
          current_rollout_status?.deployment?.status ?? ""
        );
      })
      .catch(() => {
        current_rollout_status = undefined;
        rollout_status_color = "yellow";
      });
  };

  $: if ($navigating) {
    updateCurrentRollout();
  }

  // update rollout status every 5 seconds
  onMount(() => {
    updateCurrentRollout();
  });

  // filter $rollouts by $rollouts.expand.project
  let these_rollouts: RolloutsResponse<Rexpand>[] = [];
  // @ts-ignore
  $: these_rollouts = $rollouts.filter((r) => r.expand?.project.id === project.id);

  let current_rollout: RolloutsResponse<Rexpand> | undefined;

  $: current_rollout = these_rollouts.find((r) => !r.endDate);

  type Ingress = {
    host: string;
    tls: boolean;
  };

  let ingresses: Ingress[] = [];

  $: if (current_rollout && current_rollout.manifest && current_rollout.manifest.spec.interfaces) {
    // @ts-ignore
    current_rollout.manifest.spec.interfaces.forEach((inf) => {
      if (inf.ingress) {
        // @ts-ignore
        inf.ingress.rules.forEach((rule) => {
          ingresses.push({ host: rule.host, tls: inf.tls });
        });
      }
    });
  }
</script>

<div class="rounded-xl border border-gray-200 ov">
  <div class="flex items-center gap-x-4 border-b border-gray-900/5 p-6">
    <div class="relative">
      {#if project.avatar}
        <img
          src={recordLogoUrl(project)}
          alt="Tuple"
          class="h-12 w-12 flex-none rounded-lg object-cover ring-1 ring-gray-900/10 p-1"
        />
      {:else}
        <img
          src={recordLogoUrl(project.expand?.blueprint)}
          alt="Tuple"
          class="h-12 w-12 flex-none rounded-lg object-cover ring-1 ring-gray-900/10 p-1"
        />
      {/if}
      <Indicator
        color="dark"
        size="xl"
        placement="top-right"
        class="text-xs font-bold text-white cursor-default"
        >{these_rollouts.length || 0}
      </Indicator>
      <Tooltip>Rollouts</Tooltip>
    </div>
    <div class="text-sm font-medium leading-6">{project.name}</div>
    <div class="relative ml-auto">
      <div class="flex justify-end">
        <Button
          color="alternative"
          on:click={() => {
            $selectedProjectId = project.id;
            goto(`/app/projects/${project.id}/overview`);
          }}
        >
          <ArrowRight class="w-5 h-5" />
        </Button>
      </div>
    </div>
  </div>
  <dl class="-my-3 divide-y divide-gray-100 px-6 py-4 text-sm leading-6">
    {#if these_rollouts.length > 0}
      <div class="flex justify-between gap-x-4 py-3">
        <dt class="">Last rollout</dt>
        <dd class=" cursor-default">
          <time datetime={formatDateTime(these_rollouts[0].startDate)}>
            {timeAgo(these_rollouts[0].startDate)}
          </time>
          <Tooltip>{formatDateTime(these_rollouts[0].startDate)}</Tooltip>
        </dd>
      </div>
    {/if}
    <div class="flex justify-between gap-x-4 py-3">
      <dt class="">Status</dt>
      <dd class="flex items-start gap-x-2">
        <Badge color={rollout_status_color} large class="cursor-default">
          <Indicator color={rollout_status_color} size="xs" class="mr-2" />
          {current_rollout_status?.deployment?.status ?? "Unknown"}
        </Badge>
      </dd>
    </div>
    <div class="flex justify-between gap-x-4 py-3">
      <dt class="">Hosts</dt>
      <dd class="items-start gap-x-2">
        {#if ingresses.length > 0}
          {#each ingresses as ingress (ingress)}
            <a
              href={(ingress.tls ? "https://" : "http://") + ingress.host}
              target="_blank"
              rel="noopener noreferrer"
              class="text-blue-500 hover:underline"
            >
              {ingress.host}
              <ExternalLink class="w-4 h-4 inline-block ml-1" />
            </a>
            <br />
          {/each}
        {:else}
          <a href={`/app/projects/${project.id}/network`} class="text-blue-500 hover:underline">
            Configure
            <Cog class="w-4 h-4 inline-block " />
          </a>
        {/if}
      </dd>
    </div>
    {#if tags}
      <div class="flex justify-between gap-x-4 py-3">
        <dt class="">
          <Tag class="w-5 h-5 inline-block" strokeWidth={2} /> Tags
        </dt>
        <dd class="items-start gap-y-2 space-x-2">
          {#each [...tags] as tag (tag)}
            <Badge color="dark" large class="cursor-default">{tag.charAt(0) + tag.slice(1)}</Badge>
          {/each}
        </dd>
      </div>
    {/if}
  </dl>
</div>
