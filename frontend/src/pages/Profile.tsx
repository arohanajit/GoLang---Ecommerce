import {
  Box,
  Button,
  Container,
  Heading,
  Stack,
  Text,
  Input,
} from '@chakra-ui/react'
import { useState } from 'react'
import { useQuery } from '@tanstack/react-query'
import { apiClient } from '../api/client'
import { User } from '../types'

const Profile = () => {
  const [isEditing, setIsEditing] = useState(false)
  const [name, setName] = useState('')
  const [email, setEmail] = useState('')

  const { data: user, isLoading } = useQuery<User>({
    queryKey: ['user'],
    queryFn: async () => {
      const response = await apiClient.get('/users/me')
      const userData = response.data.data
      setName(userData.name)
      setEmail(userData.email)
      return userData
    },
  })

  const handleSave = async () => {
    try {
      await apiClient.put('/users/me', { name, email })
      setIsEditing(false)
    } catch (error) {
      console.error('Failed to update profile:', error)
    }
  }

  if (isLoading) return <div>Loading...</div>
  if (!user) return <div>Please log in to view your profile</div>

  return (
    <Container maxW="container.sm" py={8}>
      <Stack gap={6}>
        <Heading>My Profile</Heading>

        <Box p={6} borderWidth={1} borderRadius="lg" shadow="sm">
          <Stack gap={4}>
            <Box>
              <Text fontWeight="bold" mb={2}>
                Name
              </Text>
              {isEditing ? (
                <Input
                  value={name}
                  onChange={(e) => setName(e.target.value)}
                />
              ) : (
                <Text>{name}</Text>
              )}
            </Box>

            <Box>
              <Text fontWeight="bold" mb={2}>
                Email
              </Text>
              {isEditing ? (
                <Input
                  value={email}
                  onChange={(e) => setEmail(e.target.value)}
                />
              ) : (
                <Text>{email}</Text>
              )}
            </Box>

            <Button
              colorScheme="blue"
              onClick={isEditing ? handleSave : () => setIsEditing(true)}
            >
              {isEditing ? 'Save Changes' : 'Edit Profile'}
            </Button>
          </Stack>
        </Box>
      </Stack>
    </Container>
  )
}

export default Profile 