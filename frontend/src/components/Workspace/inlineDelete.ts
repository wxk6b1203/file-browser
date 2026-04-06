export function buildInlineDeletePaths(targetPaths: string[], orderedPaths: string[]) {
  const targets = targetPaths.filter(Boolean)
  if (targets.length === 0) return []

  const targetSet = new Set(targets)
  const orderedTargets = orderedPaths.filter((path) => targetSet.has(path))
  return orderedTargets.length > 0 ? orderedTargets : targets
}

export function removeInlineDeletePath(paths: string[], targetPath: string) {
  return paths.filter((path) => path !== targetPath)
}
