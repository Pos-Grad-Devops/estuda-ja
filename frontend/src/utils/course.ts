const PALETTE = ['#7c3aed', '#db2777', '#2563eb', '#0891b2', '#16a34a', '#d97706']

export function courseHue(id: number) {
  return PALETTE[id % PALETTE.length]
}

export function courseInitials(titulo: string) {
  const parts = titulo.trim().split(/\s+/).filter(Boolean)
  if (parts.length === 0) return 'EJ'
  if (parts.length === 1) return parts[0].slice(0, 2).toUpperCase()
  return `${parts[0][0]}${parts[1][0]}`.toUpperCase()
}
