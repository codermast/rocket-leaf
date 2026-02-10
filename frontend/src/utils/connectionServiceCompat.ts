import { Call } from '@wailsio/runtime'
import * as ConnectionService from '../../bindings/rocket-leaf/internal/service/connectionservice'
import type { Connection } from '../../bindings/rocket-leaf/internal/model/models'

const ADD_CONNECTION_METHOD = 'rocket-leaf/internal/service.ConnectionService.AddConnection'
const UPDATE_CONNECTION_METHOD = 'rocket-leaf/internal/service.ConnectionService.UpdateConnection'

export interface ConnectionPayload {
  name: string
  env: string
  nameServer: string
  timeoutSec: number
  enableACL: boolean
  accessKey: string
  secretKey: string
  remark: string
}

const getErrorMessage = (error: unknown): string => {
  if (error instanceof Error) return error.message
  return String(error)
}

const isLegacyAddConnectionMismatch = (error: unknown): boolean => {
  const message = getErrorMessage(error)
  return message.includes('ConnectionService.AddConnection expects 5 arguments')
}

const isLegacyUpdateConnectionMismatch = (error: unknown): boolean => {
  const message = getErrorMessage(error)
  return message.includes('ConnectionService.UpdateConnection expects 6 arguments')
}

const assertLegacyBackendAclUnsupported = (payload: ConnectionPayload) => {
  if (!payload.enableACL) return
  throw new Error('当前后端版本不支持 ACL 参数，请重启应用后重试')
}

export const addConnectionCompat = async (payload: ConnectionPayload): Promise<Connection | null> => {
  try {
    return await ConnectionService.AddConnection(
      payload.name,
      payload.env,
      payload.nameServer,
      payload.timeoutSec,
      payload.enableACL,
      payload.accessKey,
      payload.secretKey,
      payload.remark
    )
  } catch (error) {
    if (!isLegacyAddConnectionMismatch(error)) {
      throw error
    }

    assertLegacyBackendAclUnsupported(payload)

    return await Call.ByName(
      ADD_CONNECTION_METHOD,
      payload.name,
      payload.env,
      payload.nameServer,
      payload.timeoutSec,
      payload.remark
    ) as Connection | null
  }
}

export const updateConnectionCompat = async (id: number, payload: ConnectionPayload): Promise<Connection | null> => {
  try {
    return await ConnectionService.UpdateConnection(
      id,
      payload.name,
      payload.env,
      payload.nameServer,
      payload.timeoutSec,
      payload.enableACL,
      payload.accessKey,
      payload.secretKey,
      payload.remark
    )
  } catch (error) {
    if (!isLegacyUpdateConnectionMismatch(error)) {
      throw error
    }

    assertLegacyBackendAclUnsupported(payload)

    return await Call.ByName(
      UPDATE_CONNECTION_METHOD,
      id,
      payload.name,
      payload.env,
      payload.nameServer,
      payload.timeoutSec,
      payload.remark
    ) as Connection | null
  }
}
