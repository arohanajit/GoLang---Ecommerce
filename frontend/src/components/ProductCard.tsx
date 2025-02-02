import {
  Box,
  Image,
  Heading,
  Text,
  Stack,
  Badge,
} from '@chakra-ui/react'
import { Link } from 'react-router-dom'
import { Product } from '../types'

interface ProductCardProps {
  product: Product
}

const ProductCard = ({ product }: ProductCardProps) => {
  return (
    <Link to={`/products/${product.id}`} style={{ textDecoration: 'none' }}>
      <Box
        borderWidth={1}
        borderRadius="lg"
        overflow="hidden"
        _hover={{ shadow: 'lg' }}
        transition="all 0.2s"
      >
        <Image
          src={product.imageUrl}
          alt={product.name}
          height="200px"
          width="100%"
          objectFit="cover"
        />

        <Stack p={4} gap={2}>
          <Heading size="md" truncate>
            {product.name}
          </Heading>

          <Text color="blue.600" fontSize="xl" fontWeight="bold">
            ${product.price.toFixed(2)}
          </Text>

          <Badge colorScheme={product.stock > 0 ? 'green' : 'red'} width="fit-content">
            {product.stock > 0 ? 'In Stock' : 'Out of Stock'}
          </Badge>

          <Text color="gray.600" truncate>
            {product.description}
          </Text>
        </Stack>
      </Box>
    </Link>
  )
}

export default ProductCard 