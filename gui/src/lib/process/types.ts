export interface ProcessDescription {
  id: string;
  name: string;
  running: boolean;
}

export interface CreateProcessResponse {
  id: string;
}

export interface ComponentStatus {
  componentId: string;
  running: boolean;
  envVars: null | Record<string, string>;
  CPUPercent: number;
  createTime: number;
  residentMemory: number;
  ports: number[];
  childrenExecutables: string[];
}

export interface ComponentDetails {
  id: string;
  name: string;
  status: ComponentStatus;
}
