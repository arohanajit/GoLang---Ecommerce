import {
  Box,
  Button,
  Container,
  FormControl,
  FormLabel,
  Heading,
  Input,
  Stack,
  Text,
  Link,
} from '@chakra-ui/react'
import { useState } from 'react'
import { Link as RouterLink, useNavigate } from 'react-router-dom'
import { apiClient } from '../api/client'

const Login = () => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const navigate = useNavigate()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)

    try {
      const response = await apiClient.post('/auth/login', {
        email,
        password,
      })

      localStorage.setItem('token', response.data.data.token)
      navigate('/')
    } catch (error) {
      console.error('Login failed:', error)
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Container maxW="container.sm" py={12}>
      <Box p={8} borderWidth={1} borderRadius="lg" boxShadow="lg">
        <Stack gap={4}>
          <Heading textAlign="center">Sign In</Heading>
          <form onSubmit={handleSubmit}>
            <Stack gap={4}>
              <FormControl>
                <FormLabel>Email</FormLabel>
                <Input
                  type="email"
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                  required
                />
              </FormControl>
              <FormControl>
                <FormLabel>Password</FormLabel>
                <Input
                  type="password"
                  value={password}
                  onChange={(e) => setPassword(e.target.value)}
                  required
                />
              </FormControl>
              <Button
                type="submit"
                colorScheme="blue"
                size="lg"
                fontSize="md"
                isLoading={isLoading}
              >
                Sign in
              </Button>
            </Stack>
          </form>
          <Text textAlign="center">
            Don't have an account?{' '}
            <Link as={RouterLink} to="/register" color="blue.500">
              Sign up
            </Link>
          </Text>
        </Stack>
      </Box>
    </Container>
  )
}

export default Login 