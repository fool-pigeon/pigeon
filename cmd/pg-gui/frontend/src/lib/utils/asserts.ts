import type { NodeMode } from "@bindings/service";

/**
 * Assert the src contains dst
 * @param src NodeMode
 * @param dst NodeMode
 * @returns boolean
 */
export function ModeContains(src: NodeMode, dst: NodeMode): boolean {
    return (src & dst) !== 0
}