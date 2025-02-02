import {
  Container,
  Grid,
  Heading,
  Stack,
  Input,
  Select,
} from '@chakra-ui/react'
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { apiClient } from '../api/client'
import { Product } from '../types'
import ProductCard from '../components/ProductCard'

const Home = () => {
  const [search, setSearch] = useState('')
  const [category, setCategory] = useState('')

  const { data: products, isLoading } = useQuery<Product[]>({
    queryKey: ['products', search, category],
    queryFn: async () => {
      const response = await apiClient.get('/products', {
        params: {
          search,
          category,
        },
      })
      return response.data.data
    },
  })

  if (isLoading) return <div>Loading...</div>

  return (
    <Container maxW="container.xl" py={8}>
      <Stack gap={8}>
        <Heading>Products</Heading>

        <Stack direction={{ base: 'column', md: 'row' }} gap={4}>
          <Input
            placeholder="Search products..."
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
          <Select
            placeholder="All Categories"
            value={category}
            onChange={(e) => setCategory(e.target.value)}
          >
            <option value="electronics">Electronics</option>
            <option value="clothing">Clothing</option>
            <option value="books">Books</option>
            <option value="home">Home & Kitchen</option>
          </Select>
        </Stack>

        <Grid
          templateColumns={{
            base: '1fr',
            md: 'repeat(2, 1fr)',
            lg: 'repeat(3, 1fr)',
          }}
          gap={6}
        >
          {products?.map((product) => (
            <ProductCard key={product.id} product={product} />
          ))}
        </Grid>
      </Stack>
    </Container>
  )
}

export default Home 