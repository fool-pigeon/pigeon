// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
// 
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

/**
 * return the path params form URL,like /a/{b}/c -> ['b']
 * @param url string
 * @returns ItertorObject<string>
 */
function ReUrlPathParams(url: string): IteratorObject<string> {
    return url.matchAll(/(?<=^|\/)\{([^/{}]+)\}/g).map(m => m[1])
}
export const regex = {
    /**
     * return the path params form URL,like /a/{b}/c -> ['b']
     * @param url string
     * @returns iter of path params
     */
    urlPathParams: ReUrlPathParams
}