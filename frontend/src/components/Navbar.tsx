import {
  Box,
  Button,
  Container,
  Flex,
  HStack,
  Heading,
  Icon,
  Link,
  Text,
} from '@chakra-ui/react'
import { Link as RouterLink } from 'react-router-dom'

const Navbar = () => {
  const isLoggedIn = !!localStorage.getItem('token')

  return (
    <Box bg="white" shadow="sm" position="sticky" top={0} zIndex={1000}>
      <Container maxW="container.xl">
        <Flex py={4} justify="space-between" align="center">
          <Link as={RouterLink} to="/" _hover={{ textDecoration: 'none' }}>
            <Heading size="lg" color="blue.600">
              E-Shop
            </Heading>
          </Link>

          <HStack spacing={8}>
            <Link as={RouterLink} to="/cart">
              Cart
            </Link>

            {isLoggedIn ? (
              <>
                <Link as={RouterLink} to="/orders">
                  Orders
                </Link>
                <Link as={RouterLink} to="/profile">
                  Profile
                </Link>
                <Button
                  onClick={() => {
                    localStorage.removeItem('token')
                    window.location.href = '/login'
                  }}
                  variant="ghost"
                >
                  Logout
                </Button>
              </>
            ) : (
              <>
                <Link as={RouterLink} to="/login">
                  Login
                </Link>
                <Button as={RouterLink} to="/register" colorScheme="blue">
                  Sign Up
                </Button>
              </>
            )}
          </HStack>
        </Flex>
      </Container>
    </Box>
  )
}

export default Navbar 