import { type Ref, toRaw } from 'vue'
import type { TabNode, TabGroupNode, TabSplitNode } from '../types'
import { genId } from '../types'

// ─── helpers ────────────────────────────────────────────────

/** Recursively strip Vue reactive proxies from a node tree */
function rawNode<T extends TabNode>(node: T): T {
  const raw = toRaw(node)
  if (raw.type === 'split') {
    return {
      ...raw,
      children: raw.children.map((c) => rawNode(toRaw(c))),
    } as T
  }
  if (raw.type === 'tabs') {
    return {
      ...raw,
      tabs: raw.tabs.map((t) => toRaw(t)),
    } as T
  }
  return raw
}

/** Deep-find a node by id */
function findNode(root: TabNode, id: string): TabNode | null {
  if (root.id === id) return root
  if (root.type === 'split') {
    for (const child of root.children) {
      const found = findNode(child, id)
      if (found) return found
    }
  }
  return null
}

/** Find the parent split node that contains child with `childId` */
function findParent(root: TabNode, childId: string): TabSplitNode | null {
  if (root.type === 'split') {
    for (const child of root.children) {
      if (child.id === childId) return root
      const found = findParent(child, childId)
      if (found) return found
    }
  }
  return null
}

function parseSizeWeight(size: string | number | undefined): number {
  if (typeof size === 'number') {
    return Number.isFinite(size) && size > 0 ? size : 0
  }
  if (typeof size !== 'string') return 0

  const trimmed = size.trim()
  if (!trimmed) return 0

  const numeric = Number.parseFloat(trimmed)
  return Number.isFinite(numeric) && numeric > 0 ? numeric : 0
}

function normalizeSizesByWeight(sizes: (string | number)[], count: number): string[] {
  if (count <= 0) return []
  if (count === 1) return ['100%']

  const weights = sizes.slice(0, count).map((s) => parseSizeWeight(s))
  const total = weights.reduce((sum, w) => sum + w, 0)
  if (total <= 0) {
    return Array.from({ length: count }, () => `${(100 / count).toFixed(4)}%`)
  }
  return weights.map((w) => `${((w / total) * 100).toFixed(4)}%`)
}

/** Replace a node inside the tree (returns new plain-object tree) */
function replaceNode(root: TabNode, targetId: string, replacement: TabNode | null): TabNode | null {
  const raw = toRaw(root)
  if (raw.id === targetId) return replacement
  if (raw.type === 'split') {
    const baseSizes = raw.sizes?.slice(0, raw.children.length) ?? []
    const newChildren: TabNode[] = []
    const preservedSlotSizes: (string | number)[] = []
    for (let i = 0; i < raw.children.length; i++) {
      const child = raw.children[i]
      if (!child) continue
      const result = replaceNode(toRaw(child), targetId, replacement)
      if (result) {
        newChildren.push(result)
        const fallbackSize = `${100 / raw.children.length}%`
        preservedSlotSizes.push(baseSizes[i] ?? fallbackSize)
      }
    }
    // If only one child remains, collapse
    if (newChildren.length === 1) return newChildren[0]!
    if (newChildren.length === 0) return null

    const nextSizes =
      newChildren.length === raw.children.length
        ? preservedSlotSizes
        : normalizeSizesByWeight(preservedSlotSizes, newChildren.length)

    return { ...raw, children: newChildren, sizes: nextSizes }
  }
  return raw
}

// ─── composable ─────────────────────────────────────────────

export function useTabTree(tree: Ref<TabNode>) {
  /** Set the active tab in a group */
  function setActive(groupId: string, tabId: string) {
    const node = findNode(tree.value, groupId)
    if (node && node.type === 'tabs') {
      node.activeId = tabId
      // Must replace tree.value with a new reference so Vue prop diffing
      // propagates the change through TabNodeRenderer → TabGroup
      tree.value = rawNode(tree.value)
    }
  }

  /** Move a tab from one group to another (or reorder within the same group) */
  function moveTab(fromGroupId: string, tabId: string, toGroupId: string, newIndex: number) {
    const from = findNode(tree.value, fromGroupId) as TabGroupNode | null
    const to = findNode(tree.value, toGroupId) as TabGroupNode | null
    if (!from || !to) return

    const tabIdx = from.tabs.findIndex((t) => t.id === tabId)
    if (tabIdx === -1) return
    const removed = from.tabs.splice(tabIdx, 1)
    const tab = toRaw(removed[0]!)
    if (!tab) return

    // Fix activeId in source group
    if (from.activeId === tabId) {
      from.activeId = from.tabs[Math.min(tabIdx, from.tabs.length - 1)]?.id ?? ''
    }

    // Insert into target (always insert raw object)
    to.tabs.splice(newIndex, 0, tab)
    to.activeId = tab.id

    // Clean up empty source group
    if (from.tabs.length === 0 && from.id !== to.id) {
      const result = replaceNode(tree.value, from.id, null)
      if (result) tree.value = rawNode(result)
    } else {
      // In-place mutation: must produce new reference
      tree.value = rawNode(tree.value)
    }
  }

  /** Split a group: pull `tabId` out into a new group in the given zone direction */
  function splitGroup(groupId: string, tabId: string, zone: 'top' | 'bottom' | 'left' | 'right') {
    const node = findNode(tree.value, groupId) as TabGroupNode | null
    if (!node) return

    const tabIdx = node.tabs.findIndex((t) => t.id === tabId)
    if (tabIdx === -1) return

    // Snapshot the tab to pull out and the remaining tabs BEFORE any mutation
    const tab = toRaw(node.tabs[tabIdx]!)
    if (!tab) return
    const remainingTabs = node.tabs.filter((_, i) => i !== tabIdx).map((t) => toRaw(t))

    // If the source group would be empty (only had 1 tab), can't split
    if (remainingTabs.length === 0) return

    // Build the new group for the pulled-out tab
    const newGroup: TabGroupNode = {
      type: 'tabs',
      id: genId('group'),
      tabs: [{ ...tab }],
      activeId: tab.id,
    }

    // Build the remaining group (deep copy tabs to avoid shared references)
    const remainingActiveId = node.activeId === tabId
      ? (remainingTabs[Math.min(tabIdx, remainingTabs.length - 1)]?.id ?? '')
      : node.activeId
    const remainingGroup: TabGroupNode = {
      type: 'tabs',
      id: genId('group'),
      tabs: remainingTabs.map((t) => ({ ...t })),
      activeId: remainingActiveId,
    }

    // Determine split layout and child order
    const layout = (zone === 'top' || zone === 'bottom') ? 'vertical' : 'horizontal'
    const first = (zone === 'top' || zone === 'left') ? newGroup : remainingGroup
    const second = (zone === 'top' || zone === 'left') ? remainingGroup : newGroup

    // Create a split node wrapping both
    const splitNode: TabSplitNode = {
      type: 'split',
      id: genId('split'),
      layout,
      children: [first, second],
      sizes: ['50%', '50%'],
    }

    // Replace the original node with the split node in the tree
    if (tree.value.id === node.id) {
      tree.value = splitNode
    } else {
      const result = replaceNode(tree.value, node.id, splitNode)
      if (result) tree.value = rawNode(result)
    }
  }

  /** Remove a tab from a group. If group becomes empty, remove the group node. */
  function removeTab(groupId: string, tabId: string) {
    const node = findNode(tree.value, groupId) as TabGroupNode | null
    if (!node) return

    const idx = node.tabs.findIndex((t) => t.id === tabId)
    if (idx === -1) return
    node.tabs.splice(idx, 1)

    if (node.activeId === tabId) {
      node.activeId = node.tabs[Math.min(idx, node.tabs.length - 1)]?.id ?? ''
    }

    if (node.tabs.length === 0) {
      const result = replaceNode(tree.value, node.id, null)
      if (result) tree.value = rawNode(result)
    } else {
      // In-place mutation: must produce new reference
      tree.value = rawNode(tree.value)
    }
  }

  /** Persist panel sizes for a split node (used by SplitPane resize-end). */
  function setSplitSizes(splitId: string, sizes: (string | number)[]) {
    const node = findNode(tree.value, splitId)
    if (!node || node.type !== 'split') return
    node.sizes = [...sizes]
    tree.value = rawNode(tree.value)
  }

  return { setActive, moveTab, splitGroup, removeTab, setSplitSizes, findNode, findParent }
}
