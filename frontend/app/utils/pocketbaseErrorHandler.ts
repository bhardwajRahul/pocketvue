export interface PocketBaseError {
  data: Record<string, any>
  message: string
  status: number
}

export interface ParsedError {
  title: string
  message: string
  field?: string
}

export interface FieldError {
  code: string
  message: string
}

export interface PocketBaseErrorData {
  data?: Record<string, FieldError> | { data?: Record<string, FieldError> }
  message?: string
  status?: number
}

export interface ClientResponseError {
  response?: {
    data?: PocketBaseErrorData
    message?: string
    status?: number
  }
  message?: string
  status?: number
}

type ErrorContext = 'login' | 'register' | 'general'

// Error message mappings
const ERROR_MESSAGES: Record<string, string> = {
  'Failed to authenticate.':
    'Invalid email or password. Please check your credentials and try again.',
  "The request doesn't satisfy the collection requirements to authenticate.":
    'Please verify your email address before logging in. Check your inbox for a verification link.',
  'Failed to create record.':
    'Registration failed. Please check your details and try again.',
  validation_not_unique:
    'This email address is already registered. Please try logging in instead.',
  validation_required: 'This field is required.',
  validation_invalid_email: 'Please enter a valid email address.',
  validation_length_out_of_range:
    'Password must be at least 8 characters long.',
  validation_values_mismatch: 'Passwords do not match.'
}

// Field-specific error messages
const FIELD_ERROR_MESSAGES: Record<string, Record<string, string>> = {
  email: {
    validation_not_unique:
      'An account with this email already exists. Please try logging in instead.',
    validation_invalid_email: 'Please enter a valid email address.'
  },
  password: {
    validation_length_out_of_range:
      'Password must be at least 8 characters long.'
  },
  passwordconfirm: {
    validation_values_mismatch: 'Passwords do not match. Please try again.'
  },
  name: {
    validation_required: 'Name is required.'
  }
}

function formatErrorMessage(message: string): string {
  return message
    .replace(/^validation_/, '')
    .replace(/_/g, ' ')
    .replace(/\b\w/g, l => l.toUpperCase())
}

function getFieldErrorMessage(
  fieldName: string,
  errorCode: string,
  errorMessage: string
): string {
  // Check global error messages
  if (ERROR_MESSAGES[errorCode]) {
    return ERROR_MESSAGES[errorCode]
  }

  // Check field-specific messages
  const fieldMessages = FIELD_ERROR_MESSAGES[fieldName.toLowerCase()]
  if (fieldMessages?.[errorCode]) {
    return fieldMessages[errorCode]
  }

  // Fallback to formatted message
  return formatErrorMessage(errorMessage)
}

function getContextualTitle(context: ErrorContext, errorType: string): string {
  const titles: Record<ErrorContext, Record<string, string>> = {
    login: { field_error: 'Login Error', default: 'Login Failed' },
    register: {
      field_error: 'Registration Error',
      default: 'Registration Failed'
    },
    general: { default: 'Error' }
  }

  return titles[context][errorType] || titles[context].default || 'Error'
}

function normalizeError(error: unknown): PocketBaseError | null {
  // Type guard for ClientResponseError
  if (
    typeof error === 'object' &&
    error !== null &&
    'response' in error &&
    'message' in error
  ) {
    const clientError = error as ClientResponseError
    if (
      clientError.response &&
      (clientError.response.data || clientError.response.message) &&
      (clientError.response.status !== undefined || clientError.status !== undefined)
    ) {
      return {
        data: (clientError.response.data as PocketBaseErrorData)?.data || {},
        message:
          clientError.response.message ||
          (clientError.response.data as PocketBaseErrorData)?.message ||
          clientError.message ||
          '',
        status: clientError.response.status || clientError.status || 0
      }
    }
  }

  // Type guard for direct PocketBaseError
  if (
    typeof error === 'object' &&
    error !== null &&
    'data' in error &&
    'message' in error &&
    'status' in error
  ) {
    return error as PocketBaseError
  }

  // Type guard for Axios-style error
  if (
    typeof error === 'object' &&
    error !== null &&
    'response' in error &&
    (error as { response?: { data?: unknown } }).response?.data
  ) {
    const axiosError = error as { response: { data: PocketBaseErrorData } }
    return {
      data: axiosError.response.data.data || {},
      message: axiosError.response.data.message || '',
      status: axiosError.response.data.status || 0
    }
  }

  return null
}

/**
 * Parse a PocketBase error and return a user-friendly error object
 */
export function parsePocketBaseError(
  error: unknown,
  context: ErrorContext = 'general'
): ParsedError {
  const pbError = normalizeError(error)

  // Handle non-PocketBase errors
  if (!pbError) {
    const errorMessage =
      (typeof error === 'object' &&
        error !== null &&
        'message' in error &&
        typeof (error as { message: unknown }).message === 'string'
        ? (error as { message: string }).message
        : null) || 'An unexpected error occurred. Please try again.'

    return {
      title: getContextualTitle(context, 'error'),
      message: errorMessage
    }
  }

  // Extract field data (handle nested structure)
  let fieldData: Record<string, FieldError> | undefined
  if (pbError.data) {
    if (
      typeof pbError.data === 'object' &&
      'data' in pbError.data &&
      typeof pbError.data.data === 'object'
    ) {
      fieldData = pbError.data.data as Record<string, FieldError>
    } else {
      fieldData = pbError.data as Record<string, FieldError>
    }
  }

  // Handle field-specific errors
  if (fieldData && Object.keys(fieldData).length > 0) {
    const firstField = Object.keys(fieldData)[0]
    if (!firstField) {
      // Fallback if no field found
      return {
        title: getContextualTitle(context, 'error'),
        message: 'An unexpected error occurred. Please try again.'
      }
    }

    const fieldError = fieldData[firstField]
    if (fieldError && fieldError.code && fieldError.message) {
      return {
        title: getContextualTitle(context, 'field_error'),
        message: getFieldErrorMessage(
          firstField,
          fieldError.code,
          fieldError.message
        ),
        field: firstField
      }
    }
  }

  // Handle general errors
  if (pbError.message) {
    return {
      title: getContextualTitle(context, 'general'),
      message: ERROR_MESSAGES[pbError.message] || pbError.message
    }
  }

  // Fallback
  return {
    title: getContextualTitle(context, 'error'),
    message: 'An unexpected error occurred. Please try again.'
  }
}

/**
 * Get just the error message
 */
export function getPocketBaseErrorMessage(
  error: unknown,
  context: ErrorContext = 'general'
): string {
  return parsePocketBaseError(error, context).message
}

/**
 * Check if error is related to email verification
 */
export function isEmailVerificationError(error: unknown): boolean {
  const pbError = normalizeError(error)
  if (!pbError) return false

  const message = pbError.message || ''
  return message.includes('collection requirements to authenticate')
}

/**
 * Check if error is related to duplicate email
 */
export function isDuplicateEmailError(error: unknown): boolean {
  const pbError = normalizeError(error)
  if (!pbError || !pbError.data) return false

  const fieldData =
    typeof pbError.data === 'object' && 'data' in pbError.data
      ? (pbError.data.data as Record<string, FieldError>)
      : (pbError.data as Record<string, FieldError>)

  return fieldData?.email?.code === 'validation_not_unique'
}
