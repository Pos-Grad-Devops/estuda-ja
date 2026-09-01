const DATE_TIME_PATTERN = /^(\d{2})\/(\d{2})\/(\d{4})(?: (\d{2}):(\d{2}))?$/

export function formatDateBR(value: string): string {
  if (!value) return '—'
  const parsed = parseDateTimeBR(value)
  if (!parsed) return value

  const day = String(parsed.getDate()).padStart(2, '0')
  const month = String(parsed.getMonth() + 1).padStart(2, '0')
  const year = parsed.getFullYear()

  if (parsed.getHours() === 0 && parsed.getMinutes() === 0) {
    return `${day}/${month}/${year}`
  }

  return formatDateTimeBR(parsed)
}

export function formatDateTimeBR(date: Date): string {
  const day = String(date.getDate()).padStart(2, '0')
  const month = String(date.getMonth() + 1).padStart(2, '0')
  const year = date.getFullYear()
  const hours = String(date.getHours()).padStart(2, '0')
  const minutes = String(date.getMinutes()).padStart(2, '0')
  return `${day}/${month}/${year} ${hours}:${minutes}`
}

export function toInputDateTimeBR(value: string): string {
  const parsed = parseDateTimeBR(value)
  if (!parsed) return ''
  return formatDateTimeBR(parsed)
}

export function parseDateTimeBR(value: string): Date | null {
  const match = value.trim().match(DATE_TIME_PATTERN)
  if (!match) return null

  const [, day, month, year, hours = '00', minutes = '00'] = match
  const parsed = new Date(
    Number(year),
    Number(month) - 1,
    Number(day),
    Number(hours),
    Number(minutes),
  )

  if (Number.isNaN(parsed.getTime())) return null
  return parsed
}

export function isValidDateTimeBR(value: string): boolean {
  return parseDateTimeBR(value) !== null
}
