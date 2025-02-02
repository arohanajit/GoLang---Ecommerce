import {
  Box,
  Container,
  Heading,
  Stack,
  Text,
  Badge,
} from '@chakra-ui/react'
import { useQuery } from '@tanstack/react-query'
import { apiClient } from '../api/client'
import { Order } from '../types'

const getStatusColor = (status: Order['status']) => {
  switch (status) {
    case 'pending':
      return 'yellow'
    case 'processing':
      return 'blue'
    case 'shipped':
      return 'purple'
    case 'delivered':
      return 'green'
    default:
      return 'gray'
  }
}

const Orders = () => {
  const { data: orders, isLoading, error } = useQuery<Order[]>({
    queryKey: ['orders'],
    queryFn: async () => {
      const response = await apiClient.get('/orders')
      return response.data.data
    },
  })

  if (isLoading) return <div>Loading...</div>
  if (error) return <div>Error loading orders</div>
  if (!orders?.length)
    return (
      <Container maxW="container.lg" py={8}>
        <Stack gap={4} align="center">
          <Heading>No Orders Yet</Heading>
          <Text>Your order history will appear here once you make a purchase.</Text>
        </Stack>
      </Container>
    )

  return (
    <Container maxW="container.lg" py={8}>
      <Heading mb={6}>Order History</Heading>
      <Stack gap={4}>
        {orders.map((order) => (
          <Box
            key={order.id}
            p={6}
            borderWidth={1}
            borderRadius="lg"
            shadow="sm"
          >
            <Stack gap={4}>
              <Box>
                <Text fontWeight="bold">Order #{order.id}</Text>
                <Text fontSize="sm" color="gray.600">
                  {new Date(order.createdAt).toLocaleDateString()}
                </Text>
              </Box>

              <Badge colorScheme={getStatusColor(order.status)} width="fit-content">
                {order.status}
              </Badge>

              <Stack gap={4}>
                {order.items.map((item) => (
                  <Box
                    key={item.product.id}
                    p={4}
                    borderWidth={1}
                    borderRadius="md"
                  >
                    <Text fontWeight="semibold">{item.product.name}</Text>
                    <Text color="gray.600">
                      Quantity: {item.quantity} × ${item.product.price.toFixed(2)}
                    </Text>
                    <Text fontWeight="bold">
                      Subtotal: $
                      {(item.quantity * item.product.price).toFixed(2)}
                    </Text>
                  </Box>
                ))}
              </Stack>

              <Box borderTopWidth={1} pt={4}>
                <Text fontSize="xl" fontWeight="bold" textAlign="right">
                  Total: ${order.total.toFixed(2)}
                </Text>
              </Box>
            </Stack>
          </Box>
        ))}
      </Stack>
    </Container>
  )
}

export default Orders 