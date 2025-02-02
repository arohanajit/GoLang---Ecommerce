import {
  Box,
  Container,
  Grid,
  Heading,
  Text,
  Button,
  Image,
  Stack,
  Badge,
  NumberInput,
  NumberInputField,
  NumberInputStepper,
  NumberIncrementStepper,
  NumberDecrementStepper,
  useToast,
} from '@chakra-ui/react'
import { useParams } from 'react-router-dom'
import { useQuery } from '@tanstack/react-query'
import { useState } from 'react'
import { apiClient } from '../api/client'
import { Product } from '../types'

const ProductDetails = () => {
  const { id } = useParams<{ id: string }>()
  const [quantity, setQuantity] = useState(1)
  const toast = useToast()

  const { data: product, isLoading } = useQuery<Product>({
    queryKey: ['product', id],
    queryFn: async () => {
      const response = await apiClient.get(`/products/${id}`)
      return response.data.data
    },
  })

  const handleAddToCart = () => {
    // TODO: Implement add to cart functionality
    toast({
      title: 'Added to cart',
      description: `${quantity} ${product?.name} added to cart`,
      status: 'success',
      duration: 3000,
    })
  }

  if (isLoading) return <div>Loading...</div>
  if (!product) return <div>Product not found</div>

  return (
    <Container maxW="container.xl" py={8}>
      <Grid templateColumns={{ base: '1fr', md: '1fr 1fr' }} gap={8}>
        <Box>
          <Image
            src={product.imageUrl}
            alt={product.name}
            borderRadius="lg"
            width="100%"
            height="500px"
            objectFit="cover"
          />
        </Box>

        <Stack gap={4}>
          <Heading size="xl">{product.name}</Heading>
          <Text fontSize="2xl" color="blue.600" fontWeight="bold">
            ${product.price.toFixed(2)}
          </Text>
          <Badge colorScheme={product.stock > 0 ? 'green' : 'red'} fontSize="md">
            {product.stock > 0 ? 'In Stock' : 'Out of Stock'}
          </Badge>
          <Text color="gray.600">{product.description}</Text>

          <Box py={4}>
            <Text mb={2}>Quantity:</Text>
            <Stack direction="row" align="center">
              <NumberInput
                value={quantity}
                onChange={(_, value) => setQuantity(value)}
                min={1}
                max={product.stock}
                width="100px"
              >
                <NumberInputField />
                <NumberInputStepper>
                  <NumberIncrementStepper />
                  <NumberDecrementStepper />
                </NumberInputStepper>
              </NumberInput>
              <Button
                colorScheme="blue"
                size="lg"
                onClick={handleAddToCart}
                isDisabled={product.stock === 0}
              >
                Add to Cart
              </Button>
            </Stack>
          </Box>

          <Box py={4}>
            <Heading size="md" mb={2}>
              Product Details
            </Heading>
            <Stack gap={2}>
              <Text>
                <strong>Category:</strong> {product.category}
              </Text>
              <Text>
                <strong>Stock:</strong> {product.stock} units
              </Text>
              <Text>
                <strong>Added:</strong>{' '}
                {new Date(product.createdAt).toLocaleDateString()}
              </Text>
            </Stack>
          </Box>
        </Stack>
      </Grid>
    </Container>
  )
}

export default ProductDetails 