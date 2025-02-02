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
  useToast,
} from '@chakra-ui/react'
import { Link, useNavigate } from 'react-router-dom'
import { useState } from 'react'
import { apiClient } from '../api/client'

const Register = () => {
  const [email, setEmail] = useState('')
  const [password, setPassword] = useState('')
  const [name, setName] = useState('')
  const [isLoading, setIsLoading] = useState(false)
  const navigate = useNavigate()
  const toast = useToast()

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setIsLoading(true)

    try {
      const response = await apiClient.post('/auth/register', {
        email,
        password,
        name,
      })

      localStorage.setItem('token', response.data.token)
      toast({
        title: 'Registration successful',
        status: 'success',
        duration: 3000,
      })
      navigate('/')
    } catch (error) {
      toast({
        title: 'Registration failed',
        description: 'Please try again',
        status: 'error',
        duration: 3000,
      })
    } finally {
      setIsLoading(false)
    }
  }

  return (
    <Container maxW="container.sm" py={8}>
      <Stack gap={6}>
        <Heading>Create an Account</Heading>
        <Box as="form" onSubmit={handleSubmit}>
          <Stack gap={4}>
            <FormControl isRequired>
              <FormLabel>Name</FormLabel>
              <Input
                type="text"
                value={name}
                onChange={(e) => setName(e.target.value)}
              />
            </FormControl>

            <FormControl isRequired>
              <FormLabel>Email</FormLabel>
              <Input
                type="email"
                value={email}
                onChange={(e) => setEmail(e.target.value)}
              />
            </FormControl>

            <FormControl isRequired>
              <FormLabel>Password</FormLabel>
              <Input
                type="password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
              />
            </FormControl>

            <Button
              type="submit"
              colorScheme="blue"
              size="lg"
              isLoading={isLoading}
            >
              Register
            </Button>
          </Stack>
        </Box>

        <Text textAlign="center">
          Already have an account?{' '}
          <Link to="/login" style={{ color: 'blue' }}>
            Login here
          </Link>
        </Text>
      </Stack>
    </Container>
  )
}

export default Register 