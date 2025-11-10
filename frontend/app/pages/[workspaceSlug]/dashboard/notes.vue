<template>
  <UDashboardPanel
    :ui="{
      body: 'bg-neutral-100 dark:bg-neutral-950'
    }"
  >
    <template #header>
      <UDashboardNavbar title="Notes" :ui="{ root: 'sm:px-4' }">
        <template #leading>
          <UDashboardSidebarCollapse :ui="{ leadingIcon: 'size-4' }" />
          <USeparator orientation="vertical" class="mr-2 h-4" />
        </template>
        <template #right>
          <UButton
            icon="i-lucide-plus"
            class="rounded-full"
            variant="soft"
            size="sm"
            @click="openAddNoteModal"
          />
        </template>
      </UDashboardNavbar>
    </template>

    <template #body>
      <div
        class="relative column-1 md:columns-2 lg:columns-4 gap-4 space-y-4
          *:break-inside-avoid-column *:will-change-transform"
      >
        <NotesCard v-for="note in notes" :key="note.id" :note="note" />
      </div>
    </template>
  </UDashboardPanel>
  <UModal
    v-model:open="addNoteModal"
    title="New Note"
    :ui="{ content: 'sm:max-w-lg' }"
  >
    <template #body>
      <NotesForm @close="addNoteModal = false" />
    </template>
  </UModal>
</template>

<script setup lang="ts">
const addNoteModal = ref(false)
const { notes, subscribeNotes, unsubscribeNotes, fetchNotes } = useNotes()

const { data } = await useAsyncData('all-notes', () => fetchNotes())

notes.value = data.value || []

onMounted(async () => {
  await subscribeNotes()
})

onUnmounted(() => {
  unsubscribeNotes()
})

const openAddNoteModal = () => {
  addNoteModal.value = true
}
</script>
