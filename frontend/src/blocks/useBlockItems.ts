import { getBlockItems, type HomeBlock } from '../api/home'
import { useAsync } from '../lib/useAsync'

/**
 * Loads the content of one homepage block from the endpoint named in the
 * layout. Each block calls this for itself, so blocks load side by side and
 * one failing does not affect the others.
 *
 * A failed request is reported as "no items". A block with no items draws
 * nothing: a missing row of products is better than an error box in the
 * middle of a shop's front page.
 */
export function useBlockItems<T>(block: HomeBlock): { loading: boolean; items: T[] } {
  const state = useAsync((signal) => getBlockItems<T>(block.endpoint, signal), block.endpoint)
  return {
    loading: state.status === 'loading',
    items: state.status === 'ready' ? state.data : [],
  }
}
