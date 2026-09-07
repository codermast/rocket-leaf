import { DriverService } from "@bindings/bridge";
import type { Capabilities, MQKind } from "@bindings/model/models";
import { required } from "./client";

export type { Capabilities, MQKind };

/** What one live connection can actually do. */
export const getCapabilities = (connID: number): Promise<Capabilities> =>
  DriverService.Capabilities(connID).then(required);
