import type { ProductSummary } from '../api/catalog'
import { ProductCard, ProductCardSkeleton } from '../components/ProductCard'
import { type BlockProps, BlockSection } from './BlockSection'
import { useBlockItems } from './useBlockItems'

/**
 * A row of product cards. Best sellers and New releases are both this
 * component: they differ only in the endpoint the layout gives them, and in
 * whether the cards are numbered.
 */
export function ProductRowBlock({ block, ranked = false }: BlockProps & { ranked?: boolean }) {
  const { loading, items: products } = useBlockItems<ProductSummary>(block)
  if (!loading && products.length === 0) return null

  return (
    <BlockSection block={block}>
      <div className="product-row">
        {loading
          ? Array.from({ length: 6 }, (_, index) => <ProductCardSkeleton key={index} />)
          : products.map((product, index) => (
              <ProductCard key={product.slug} product={product} rank={ranked ? index + 1 : undefined} />
            ))}
      </div>
    </BlockSection>
  )
}

/** Best sellers: the same row, with rank numbers on the cards. */
export function BestSellersBlock({ block }: BlockProps) {
  return <ProductRowBlock block={block} ranked />
}
