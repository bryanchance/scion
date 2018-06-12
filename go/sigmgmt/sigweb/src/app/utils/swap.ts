/* tslint:disable */
declare interface Array<T> {
    swap(a: number, b: number);
}

/**
 * Swap array elements position
 *
 * @param {number} a - first index
 * @param {number} b - second index
 *
 * return {object} this
 */
Array.prototype.swap = function (a: number, b: number) {
    if (a < 0 || a >= this.length || b < 0 || b >= this.length) {
        return
    }

    const temp = this[a]
    this[a] = this[b]
    this[b] = temp
}
