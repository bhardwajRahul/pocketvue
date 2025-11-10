export default defineNuxtRouteMiddleware(async to => {
  // Handle root path redirect for logged-in users
  if (to.path === '/' && !to.params.workspaceSlug) {
    const { isLoggedIn } = useAuth()
    if (isLoggedIn.value) {
      const { fetchWorkspaces } = useWorkspaces()
      const { redirectToDefaultWorkspace } = useWorkspaceNavigation()

      await fetchWorkspaces()
      return redirectToDefaultWorkspace()
    }
    return
  }

  // Only validate workspace routes
  if (!to.params.workspaceSlug) {
    return
  }

  const { ensureWorkspacesLoaded, findWorkspaceBySlug, redirectToDefaultWorkspace } =
    useWorkspaceNavigation()
  const { fetchWorkspaces } = useWorkspaces()

  // Ensure workspaces are loaded
  await ensureWorkspacesLoaded()

  // Validate workspace exists
  let workspace = findWorkspaceBySlug(to.params.workspaceSlug as string)

  // Try refreshing if not found (stale cache)
  if (!workspace) {
    await fetchWorkspaces()
    workspace = findWorkspaceBySlug(to.params.workspaceSlug as string)
  }

  if (!workspace) {
    throw createError({
      statusCode: 404,
      statusMessage: 'Workspace not found'
    })
  }
})
