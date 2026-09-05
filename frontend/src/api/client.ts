import { getToken } from '../auth/auth'
import type { Role, User } from '../auth/auth'

// VITE_API_URL é embutido no build (Vite). Mudar a sessão = alterar env e rebuild.
export const API_URL = import.meta.env.VITE_API_URL ?? 'http://localhost:8080'

/** Converte base HTTP(S) da API em URL WebSocket do chat da aula (`?token=` na query). */
export function buildAulaChatWsUrl(aulaId: number, token: string): string {
  const wsBase = API_URL.replace(/^https:/i, 'wss:').replace(/^http:/i, 'ws:')
  const path = `/api/v1/aulas/${aulaId}/chat/ws`
  return `${wsBase}${path}?token=${encodeURIComponent(token)}`
}

export type Curso = {
  id: number
  titulo: string
  descricao: string
  created_at: string
  updated_at: string
}

export type Aula = {
  id: number
  curso_id: number
  titulo: string
  descricao: string
  agendada_em: string
  status: 'agendada' | 'ao_vivo' | 'encerrada'
  created_at: string
  updated_at: string
  curso?: Curso
}

export type Aluno = {
  id: number
  nome: string
  email: string
  created_at: string
  updated_at: string
}

export type UserInput = {
  nome: string
  email: string
  password?: string
  role: Role
}

export type AulaVod = {
  aula_id: number
  status: 'publicado' | 'rascunho'
  content_type: string
  size_bytes: number
  updated_at: string
}

export type VodPlayback = {
  playback_url: string
  expires_at: string
  expires_in_seconds: number
}

export type LiveStatus = 'inativa' | 'agendada' | 'ao_vivo' | 'encerrada'

export type LiveModo = 'stub' | 'ivs'

export type AulaLive = {
  aula_id: number
  status: LiveStatus
  modo: LiveModo
  iniciada_em: string | null
  encerrada_em: string | null
}

export type LivePlayback = {
  aula_id: number
  modo: LiveModo
  protocolo: 'hls' | null
  player: 'ivs' | null
  playback_url: string | null
  mensagem?: string
}

export type LiveIngest = {
  aula_id: number
  modo: LiveModo
  ingest_server: string | null
  stream_key: string | null
  observacao?: string
  mensagem?: string
}

export type CertificadoStatus = 'valido' | 'invalidado'

export type CertificadoResumo = {
  id: number
  status: CertificadoStatus
  emitido_em: string
  aluno_nome: string
  curso_titulo: string
  user_id?: number
}

export type CertificadoStatusResponse = {
  curso_id: number
  user_id: number
  elegivel: boolean
  certificado: CertificadoResumo | null
}

export type CertificadoElegibilidadeItem = {
  user_id: number
  user_nome: string
  created_at: string
}

export type CertificadosListResponse = {
  curso_id: number
  elegibilidades: CertificadoElegibilidadeItem[]
  certificados: CertificadoResumo[]
}

export type CertificadoElegibilidadeResponse = {
  curso_id: number
  user_id: number
  elegivel: boolean
}

export type CertificadoInvalidarResponse = {
  id: number
  status: CertificadoStatus
  curso_id: number
  user_id: number
}

async function parseError(response: Response): Promise<Error> {
  if (response.status === 401) {
    return new Error('sessão expirada, faça login novamente')
  }
  const body = await response.json().catch(() => ({}))
  return new Error((body as { error?: string }).error ?? `Erro ${response.status}`)
}

async function request<T>(path: string, options?: RequestInit): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(options?.headers as Record<string, string> | undefined),
  }
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  const response = await fetch(`${API_URL}${path}`, {
    ...options,
    headers,
  })

  if (!response.ok) {
    throw await parseError(response)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json()
}

/** Multipart sem Content-Type fixo (boundary do browser). */
async function requestForm<T>(path: string, formData: FormData, method: string): Promise<T> {
  const token = getToken()
  const headers: Record<string, string> = {}
  if (token) {
    headers.Authorization = `Bearer ${token}`
  }

  const response = await fetch(`${API_URL}${path}`, {
    method,
    headers,
    body: formData,
  })

  if (!response.ok) {
    throw await parseError(response)
  }

  if (response.status === 204) {
    return undefined as T
  }

  return response.json()
}

export const api = {
  auth: {
    login: (email: string, password: string) =>
      request<{ token: string; user: User }>('/api/v1/auth/login', {
        method: 'POST',
        body: JSON.stringify({ email, password }),
      }),
    me: () => request<User>('/api/v1/auth/me'),
  },
  cursos: {
    list: () => request<Curso[]>('/api/v1/cursos'),
    create: (data: Pick<Curso, 'titulo' | 'descricao'>) =>
      request<Curso>('/api/v1/cursos', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: number, data: Pick<Curso, 'titulo' | 'descricao'>) =>
      request<Curso>(`/api/v1/cursos/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    remove: (id: number) => request<void>(`/api/v1/cursos/${id}`, { method: 'DELETE' }),
  },
  aulas: {
    list: (cursoId?: number) =>
      request<Aula[]>(`/api/v1/aulas${cursoId ? `?curso_id=${cursoId}` : ''}`),
    create: (data: Pick<Aula, 'curso_id' | 'titulo' | 'descricao' | 'agendada_em' | 'status'>) =>
      request<Aula>('/api/v1/aulas', { method: 'POST', body: JSON.stringify(data) }),
    update: (
      id: number,
      data: Pick<Aula, 'curso_id' | 'titulo' | 'descricao' | 'agendada_em' | 'status'>,
    ) => request<Aula>(`/api/v1/aulas/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    remove: (id: number) => request<void>(`/api/v1/aulas/${id}`, { method: 'DELETE' }),
  },
  vod: {
    get: async (aulaId: number): Promise<AulaVod | null> => {
      const token = getToken()
      const headers: Record<string, string> = { 'Content-Type': 'application/json' }
      if (token) {
        headers.Authorization = `Bearer ${token}`
      }
      const response = await fetch(`${API_URL}/api/v1/aulas/${aulaId}/vod`, { headers })
      if (response.status === 404) {
        return null
      }
      if (!response.ok) {
        throw await parseError(response)
      }
      return response.json()
    },
    getPlayback: (aulaId: number) =>
      request<VodPlayback>(`/api/v1/aulas/${aulaId}/vod/playback`),
    upload: (aulaId: number, file: File) => {
      const formData = new FormData()
      formData.append('file', file)
      return requestForm<AulaVod>(`/api/v1/aulas/${aulaId}/vod`, formData, 'PUT')
    },
    remove: (aulaId: number) =>
      request<{ message?: string }>(`/api/v1/aulas/${aulaId}/vod`, { method: 'DELETE' }),
  },
  live: {
    get: (aulaId: number) => request<AulaLive>(`/api/v1/aulas/${aulaId}/live`),
    schedule: (aulaId: number) =>
      request<AulaLive>(`/api/v1/aulas/${aulaId}/live/schedule`, { method: 'POST' }),
    cancel: (aulaId: number) =>
      request<AulaLive>(`/api/v1/aulas/${aulaId}/live/cancel`, { method: 'POST' }),
    start: (aulaId: number) =>
      request<AulaLive>(`/api/v1/aulas/${aulaId}/live/start`, { method: 'POST' }),
    stop: (aulaId: number) =>
      request<AulaLive>(`/api/v1/aulas/${aulaId}/live/stop`, { method: 'POST' }),
    getPlayback: (aulaId: number) =>
      request<LivePlayback>(`/api/v1/aulas/${aulaId}/live/playback`),
    getIngest: (aulaId: number) =>
      request<LiveIngest>(`/api/v1/aulas/${aulaId}/live/ingest`),
  },
  alunos: {
    list: () => request<Aluno[]>('/api/v1/alunos'),
    create: (data: Pick<Aluno, 'nome' | 'email'>) =>
      request<Aluno>('/api/v1/alunos', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: number, data: Pick<Aluno, 'nome' | 'email'>) =>
      request<Aluno>(`/api/v1/alunos/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    remove: (id: number) => request<void>(`/api/v1/alunos/${id}`, { method: 'DELETE' }),
  },
  users: {
    list: () => request<User[]>('/api/v1/users'),
    create: (data: UserInput & { password: string }) =>
      request<User>('/api/v1/users', { method: 'POST', body: JSON.stringify(data) }),
    update: (id: number, data: UserInput) =>
      request<User>(`/api/v1/users/${id}`, { method: 'PUT', body: JSON.stringify(data) }),
    remove: (id: number) => request<void>(`/api/v1/users/${id}`, { method: 'DELETE' }),
  },
  certificados: {
    getStatus: (cursoId: number, userId?: number) => {
      const qs = userId != null ? `?user_id=${userId}` : ''
      return request<CertificadoStatusResponse>(`/api/v1/cursos/${cursoId}/certificado${qs}`)
    },
    downloadPdf: async (cursoId: number): Promise<void> => {
      const token = getToken()
      const headers: Record<string, string> = {}
      if (token) {
        headers.Authorization = `Bearer ${token}`
      }
      const response = await fetch(`${API_URL}/api/v1/cursos/${cursoId}/certificado/pdf`, {
        headers,
      })
      if (!response.ok) {
        throw await parseError(response)
      }
      const blob = await response.blob()
      const url = URL.createObjectURL(blob)
      const anchor = document.createElement('a')
      anchor.href = url
      anchor.download = `certificado-curso-${cursoId}.pdf`
      document.body.appendChild(anchor)
      anchor.click()
      anchor.remove()
      URL.revokeObjectURL(url)
    },
    marcarElegibilidade: (cursoId: number, userId: number) =>
      request<CertificadoElegibilidadeResponse>(
        `/api/v1/cursos/${cursoId}/certificados/elegibilidade`,
        { method: 'PUT', body: JSON.stringify({ user_id: userId }) },
      ),
    list: (cursoId: number) =>
      request<CertificadosListResponse>(`/api/v1/cursos/${cursoId}/certificados`),
    invalidar: (cursoId: number, certId: number) =>
      request<CertificadoInvalidarResponse>(
        `/api/v1/cursos/${cursoId}/certificados/${certId}/invalidar`,
        { method: 'POST' },
      ),
  },
}
