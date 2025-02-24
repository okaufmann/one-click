<script lang="ts">
    import NewVolume from "$lib/components/volumes/NewVolume.svelte";
    import { client } from "$lib/pocketbase";
  import type { RolloutsRecord, RolloutsResponse } from "$lib/pocketbase/generated-types";
  import { currentRollout, rollouts, type Rexpand, updateDataStores, UpdateFilterEnum } from "$lib/stores/data";
  import {
    Accordion,
    AccordionItem,
    Button,
    Heading,
    Input,
    Label,
    Modal,
    P,
    Toggle
  } from "flowbite-svelte";
    import { ExclamationCircleOutline } from "flowbite-svelte-icons";
  import { HardDrive, Plus } from "lucide-svelte";
    import toast from "svelte-french-toast";

  interface Volume {
    id: string;
    name: string;
    mountPath: string;
    size: string;
    storageClass: string;
  }

  export let modal: boolean;
  let deleteModal: boolean = false;
  let selectedVolumeId: string = "";

  let volumes: Volume[] = [];

  $: {
    parseManifestsToVolumes($currentRollout);
  }

  function parseManifestsToVolumes(rollout: RolloutsResponse<Rexpand> | undefined) {
    volumes = [];
    if (rollout?.manifest && rollout.manifest.spec && rollout.manifest.spec.volumes) {
      rollout?.manifest.spec.volumes.forEach((v: any, index: number) => {
        const volumeId = `${rollout?.id}_${index}`;

        let newVolume: Volume = {
          id: volumeId,
          name: v.name,
          mountPath: v.mountPath || "",
          size: v.size || "",
          storageClass: v.storageClass || ""
        };
        volumes.push(newVolume);
      });
    }
  }

  function handleDeleteClick(volumeId: string) {
    selectedVolumeId = volumeId;
    deleteModal = true;
  }

  async function handleDelete() {
    const volumeIndex = volumes.findIndex((volume) => volume.id === selectedVolumeId);
    if (volumeIndex === -1) {
      return; // volume not found, do nothing
    }

    volumes.splice(volumeIndex, 1); // Remove the volume from the array

    if (!$currentRollout) {
      toast.error("No rollout selected");
      return;
    }

    if (!$currentRollout.manifest) {
      toast.error("No manifest found");
      return;
    }

    // Remove the volume from the manifest
    $currentRollout.manifest.spec.volumes.splice(volumeIndex, 1);

    // Save the manifest
    try {
      await updateManifest($currentRollout.manifest);
      toast.success("Volume deleted successfully");
    } catch (error) {
      toast.error("Failed to delete volume");
      console.error(error);
    }
  }

  async function handleInputSave(id: string) {
    const volumeIndex = volumes.findIndex((volume) => volume.id === id);
    if (!$currentRollout) {
      toast.error("No rollout selected");
      return;
    }

    const updatedVolume = volumes[volumeIndex];

    if (!updatedVolume.name) {
      toast.error("Volume name is required");
      return;
    }

    if (!updatedVolume.mountPath) {
      toast.error("Mount path is required");
      return;
    }

    // Find the index of the current volume based on its unique identifier (id)
    // @ts-ignore
    const currentVolumeIndex = $currentRollout.manifest.spec.volumes.findIndex(
      (volume: any) => volume.id === updatedVolume.id
    );

    if (currentVolumeIndex === -1) {
      // Check if there's another volume with the same name or mount path. Make sure it's not the same volume
      // @ts-ignore
      const existingVolume = $currentRollout.manifest.spec.volumes.find(
        (volume: any) =>
          (volume.name === updatedVolume.name || volume.mountPath === updatedVolume.mountPath) &&
          volume.id !== updatedVolume.id
      );

      // exclude if the volume is the same as the updated volume
      // and if there is only one volume
      // @ts-ignore
      if (existingVolume && $currentRollout.manifest.spec.volumes.length > 1) {
        toast.error("A volume with the same name or mount path already exists");
        return;
      }
    }

    // Update the volume in $currentRollout
    // @ts-ignore
    const rolloutVolumeIndex = $currentRollout.manifest.spec.volumes.findIndex(
      (volume: any) => volume.name === updatedVolume.name || volume.mountPath === updatedVolume.mountPath
    );

    if (rolloutVolumeIndex !== -1) {
      // @ts-ignore
      $currentRollout.manifest.spec.volumes[rolloutVolumeIndex] = updatedVolume;
    }

    // update the manifest
    if (!$currentRollout) {
      toast.error("No rollout selected");
      return;
    }

    // @ts-ignore
    if (!$currentRollout.manifest) {
      toast.error("No manifest found");
      return;
    }

    // @ts-ignore
    await updateManifest($currentRollout.manifest);

    toast.success("Volume updated successfully");


  }

  async function updateManifest(manifest: any) {
    try {
      if (!$currentRollout) {
        toast.error("No rollout selected");
        return;
      }
      const data: RolloutsRecord = {
        manifest: manifest,
        startDate: $currentRollout?.startDate,
        endDate: "",
        project: $currentRollout?.project,
        user: client.authStore.model?.id
      };

      client
        .collection("rollouts")
        .create(data)
        .then((res) => {
          updateDataStores({
            filter: UpdateFilterEnum.ALL,
            projectId: $currentRollout?.project
          });
        });

      // Update the rollout in the store

      // update the $rollouts store
      rollouts.update((rollouts) => {
        const rolloutIndex = rollouts.findIndex((r) => r.id === $currentRollout?.id);
        if (rolloutIndex !== -1) {
          rollouts[rolloutIndex] = {
            ...rollouts[rolloutIndex],
            manifest: manifest
          };
        }
        return rollouts;
      });

      $currentRollout.manifest = manifest;

      // Update the rollout in the store
    } catch (error) {
      console.error("Failed to update manifest:", error);
      toast.error("Failed to update interface.");
    }
  }

</script>

<div class="flex items-start justify-between"></div>

<div class="flex items-start justify-between">
  <div class="flex flex-col">
    <Heading tag="h2">Volumes</Heading>
    <P class="text-gray-500 dark:text-gray-400 text-sm">Persistent volumes for your rollout.</P>
  </div>
  <Button color="primary" size="xs" class="whitespace-nowrap self-start" on:click={() => {
    modal = true;
  }}>
    <Plus class="w-4 h-4 mr-2" />
    New volume
  </Button>
</div>

<Accordion class="gap-2 grid mt-10" flush>
  {#key $rollouts}
    {#each volumes as volume, v (volume.id)}
      <AccordionItem class="rounded-lg">
        <div slot="header" class="flex">
          <div class="ring-1 p-2 rounded-lg ring-gray-500 mr-2 flex items-center justify-center">
            <HardDrive class="w-4 h-4" />
          </div>
          <span class="pt-1"
            >{volume.name}
            <span class="font-normal text-sm">{volume.size} {volume.mountPath}</span></span
          >
        </div>
        <div class="">
          <table class="min-w-full divide-y divide-gray-300 dark:divide-gray-600">
            <tbody class="divide-y divide-gray-200 dark:divide-gray-600">
              <tr class="transition-all hover:bg-gray-50 dark:hover:bg-gray-800">
                <td class="whitespace-nowrap py-4 pl-4 pr-3 text-xs font-medium sm:pl-6">
                  <Heading tag="h5">Details</Heading>
                  <P class="text-gray-500 dark:text-gray-400 text-xs">Details of your Volume.</P>
                </td><td class="whitespace-nowrap px-3 py-4 text-xs space-y-2">
                  <Label for="tag" class="block ">Volume name *</Label>
                  <Input
                    id="name"
                    size="sm"
                    type="text"
                    bind:value={volume.name}
                    placeholder="Enter the name of your volume"
                    class=""
                  />
                  <Label for="tag" class="block ">Mount path *</Label>
                  <Input
                    id="mountPath"
                    size="sm"
                    type="text"
                    bind:value={volume.mountPath}
                    placeholder="/data"
                    class=""
                  />
                </td>
              </tr>
              <tr class="transition-all hover:bg-gray-50 dark:hover:bg-gray-800">
                <td class="whitespace-nowrap py-4 pl-4 pr-3 text-xs font-medium sm:pl-6">
                  <Heading tag="h5">Immutable Configuration</Heading>
                  <P class="text-gray-500 dark:text-gray-400 text-xs">
                    These values cannot be changed after creation.
                  </P>
                </td><td class="whitespace-nowrap px-3 py-4 text-xs space-y-2">

                  <Label for="tag" class="block ">Size</Label>
                  <Input
                    id="size"
                    size="sm"
                    type="text"
                    bind:value={volume.size}
                    placeholder="1Gi"
                    class=""
                    disabled
                  />
                  <Label for="tag" class="block ">Storage class</Label>
                  <Input
                    id="storageClass"
                    size="sm"
                    type="text"
                    bind:value={volume.storageClass}
                    placeholder="standard"
                    class=""
                    disabled
                  />
                </td>
              </tr>
            </tbody>
          </table>
          <!-- Reset & Save Button bottom right -->
          <div class="flex justify-end mt-4 p-4">
            <Button
              color="red"
              class="whitespace-nowrap self-start mr-2"
              on:click={() => handleDeleteClick(volume.id)}
            >
              Delete
            </Button>

            <Button
              color="primary"
              class="whitespace-nowrap self-start"
              on:click={() => handleInputSave(volume.id)}
            >
              Save
            </Button>
          </div>
        </div>
      </AccordionItem>
    {/each}
  {/key}
</Accordion>


<Modal bind:open={deleteModal} size="xs" autoclose>
  <div class="text-center">
    <ExclamationCircleOutline class="mx-auto mb-4 text-gray-400 w-12 h-12 dark:text-gray-200" />
    <h3 class="mb-5 text-lg font-normal text-gray-500 dark:text-gray-400">
      Are you sure you want to delete this interface?
    </h3>
    <Button color="red" class="me-2" on:click={() => handleDelete()}>Yes, I'm sure</Button>
    <Button color="alternative">No, cancel</Button>
  </div>
</Modal>

<div>
  <Modal bind:open={modal} size="xs" autoclose={false} class="w-full">
    <NewVolume bind:modal />
  </Modal>
</div>
