import {
  Box,
  Button,
  Container,
  Heading,
  Stack,
  HStack,
  Text,
  Image,
} from '@chakra-ui/react'
import { useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { CartItem } from '../types'

const Cart = () => {
  // This would typically come from a global state management solution
  const [cartItems, setCartItems] = useState<CartItem[]>([])
  const navigate = useNavigate()

  const total = cartItems.reduce(
    (sum, item) => sum + item.product.price * item.quantity,
    0
  )

  const handleUpdateQuantity = (productId: string, newQuantity: number) => {
    setCartItems((items) =>
      items.map((item) =>
        item.product.id === productId
          ? { ...item, quantity: newQuantity }
          : item
      )
    )
  }

  const handleRemoveItem = (productId: string) => {
    setCartItems((items) =>
      items.filter((item) => item.product.id !== productId)
    )
  }

  const handleCheckout = () => {
    // TODO: Implement checkout functionality
    console.log('Proceeding to checkout')
  }

  if (cartItems.length === 0) {
    return (
      <Container maxW="container.lg" py={8}>
        <Stack gap={4} align="center">
          <Heading>Your Cart is Empty</Heading>
          <Text>Add some products to your cart to see them here.</Text>
          <Button colorScheme="blue" onClick={() => navigate('/')}>
            Continue Shopping
          </Button>
        </Stack>
      </Container>
    )
  }

  return (
    <Container maxW="container.lg" py={8}>
      <Heading mb={6}>Shopping Cart</Heading>
      <Stack gap={4} align="stretch">
        {cartItems.map((item) => (
          <Box
            key={item.product.id}
            p={4}
            borderWidth={1}
            borderRadius="lg"
            shadow="sm"
          >
            <HStack gap={4}>
              <Image
                src={item.product.imageUrl}
                alt={item.product.name}
                boxSize="100px"
                objectFit="cover"
                borderRadius="md"
              />
              <Stack flex={1} gap={1}>
                <Text fontSize="lg" fontWeight="semibold">
                  {item.product.name}
                </Text>
                <Text color="gray.600">${item.product.price.toFixed(2)}</Text>
              </Stack>
              <HStack>
                <Button
                  size="sm"
                  onClick={() =>
                    handleUpdateQuantity(item.product.id, item.quantity - 1)
                  }
                  disabled={item.quantity <= 1}
                >
                  -
                </Button>
                <Text>{item.quantity}</Text>
                <Button
                  size="sm"
                  onClick={() =>
                    handleUpdateQuantity(item.product.id, item.quantity + 1)
                  }
                  disabled={item.quantity >= item.product.stock}
                >
                  +
                </Button>
              </HStack>
              <Text fontWeight="bold">
                ${(item.product.price * item.quantity).toFixed(2)}
              </Text>
              <Button
                onClick={() => handleRemoveItem(item.product.id)}
                variant="ghost"
                colorScheme="red"
              >
                Remove
              </Button>
            </HStack>
          </Box>
        ))}

        <Box borderBottomWidth={1} my={4} borderColor="gray.200" />

        <Box alignSelf="flex-end">
          <HStack gap={4} justify="space-between" width="300px">
            <Text fontSize="lg">Total:</Text>
            <Text fontSize="xl" fontWeight="bold">
              ${total.toFixed(2)}
            </Text>
          </HStack>
          <Button
            colorScheme="blue"
            size="lg"
            width="full"
            mt={4}
            onClick={handleCheckout}
          >
            Proceed to Checkout
          </Button>
        </Box>
      </Stack>
    </Container>
  )
}

export default Cart 