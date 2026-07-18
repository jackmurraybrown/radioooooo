import * as v from 'valibot'
import { reactive } from 'vue'

// ⋆˙⟡ lightweight form validation — clears + re-runs valibot on each call
export function useValidation(schema: v.GenericSchema) {
  const errors = reactive<Record<string, string>>({})

  function validate(data: unknown): boolean {
    Object.keys(errors).forEach(k => delete errors[k])
    const result = v.safeParse(schema, data)
    if (!result.success) {
      for (const issue of result.issues) {
        const key = issue.path?.[0]?.key as string
        if (key && !errors[key]) errors[key] = issue.message
      }
    }
    return result.success
  }

  return { errors, validate }
}
