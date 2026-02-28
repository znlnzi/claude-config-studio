export interface GlobalConfig {
  claudeHome: string;
  claudeMd: string;
  settings: string;
  hasClaudeMd: boolean;
  hasSettings: boolean;
}

export interface ProjectConfig {
  path: string;
  claudeMd: string;
  settings: string;
  mcpConfig: string;
  hasClaudeMd: boolean;
  hasSettings: boolean;
  hasMcp: boolean;
}

export interface ProjectInfo {
  name: string;
  path: string;
  hasClaudeMd: boolean;
  hasSettings: boolean;
  hasMcp: boolean;
  hasHooks: boolean;
  hasCommands: boolean;
  hasAgents: boolean;
  hasSkills: boolean;
  lastModified: string;
  configCount: number;
}

export interface GlobalStats {
  hasGlobalClaudeMd: boolean;
  hasGlobalSettings: boolean;
  hasLspConfig: boolean;
  globalAgentCount: number;
  globalCommandCount: number;
  projectCount: number;
  enabledPluginCount: number;
}

export type NavItem = {
  id: string;
  label: string;
  icon: string;
  children?: NavItem[];
};
