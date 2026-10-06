// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

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
/**
 * Assert the src path contains dst,AKA src path is dst's parent
 * @param src string
 * @param dst string
 */
function PathContains(src: string, dst: string): boolean {
    const srcSlice = src.split(/[\\/]/)
    const dstSlice = dst.split(/[\\/]/)
    for (let i = 0; i < srcSlice.length; i++) {
        if (dstSlice[i] !== srcSlice[i]) {
            return false
        }
    }
    return true
}

export const assertPath = {
    /**
     * Assert the src path contains dst,AKA src path is dst's parent
     * @param src string
     * @param dst string
     */
    contains: PathContains
}