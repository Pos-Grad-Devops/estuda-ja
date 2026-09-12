import { describe, expect, it } from 'vitest'
import {
  canManageAlunos,
  canManageAulas,
  canManageCertificados,
  canManageCursos,
  canManageUsers,
} from './auth'
import type { Role } from './auth'

const roles: Role[] = ['admin', 'professor', 'aluno']

describe('RBAC na UI', () => {
  it('só admin gerencia cursos, alunos, usuários e certificados', () => {
    for (const role of roles) {
      const isAdmin = role === 'admin'
      expect(canManageCursos(role)).toBe(isAdmin)
      expect(canManageAlunos(role)).toBe(isAdmin)
      expect(canManageUsers(role)).toBe(isAdmin)
      expect(canManageCertificados(role)).toBe(isAdmin)
    }
  })

  it('admin e professor gerenciam aulas; aluno não', () => {
    expect(canManageAulas('admin')).toBe(true)
    expect(canManageAulas('professor')).toBe(true)
    expect(canManageAulas('aluno')).toBe(false)
  })
})
