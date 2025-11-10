import {
  findWorkspaceBySlug,
  getDefaultWorkspace,
  getWorkspacePath
} from '~/utils/workspace'

export const useWorkspaceNavigation = () => {
  const { workspaces, fetchWorkspaces } = useWorkspaces()
  const { getLastUsedWorkspace } = useWorkspacePreferences()

  const ensureWorkspacesLoaded = async () => {
    if (workspaces.value.length === 0) {
      await fetchWorkspaces()
    }
  }

  const redirectToWorkspace = (workspaceSlug: string, subPath = '') => {
    return navigateTo(getWorkspacePath(workspaceSlug, subPath))
  }

  const redirectToDefaultWorkspace = async () => {
    await ensureWorkspacesLoaded()

    if (workspaces.value.length === 0) {
      return navigateTo('/new-workspace')
    }

    const defaultWorkspace = getDefaultWorkspace(
      workspaces.value,
      getLastUsedWorkspace()
    )

    if (defaultWorkspace) {
      return redirectToWorkspace(defaultWorkspace.slug)
    }

    return navigateTo('/new-workspace')
  }

  return {
    ensureWorkspacesLoaded,
    findWorkspaceBySlug: (slug: string) =>
      findWorkspaceBySlug(workspaces.value, slug),
    getDefaultWorkspace: () =>
      getDefaultWorkspace(workspaces.value, getLastUsedWorkspace()),
    redirectToWorkspace,
    redirectToDefaultWorkspace
  }
}
