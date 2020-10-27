export interface Archive {
  uid: uuid
  kind: string
  namespace: string
  name: string
  start_time: string
  finish_time: string
}

export interface ArchiveDetail extends Archive {
  yaml: any
}
