from transformers import AutoModelForCausalLM, AutoTokenizer

# Загрузка модели и токенизатора
model_name = "Vikhrmodels/Vikhr-Llama-3.2-1B-instruct"
model = AutoModelForCausalLM.from_pretrained(model_name)
tokenizer = AutoTokenizer.from_pretrained(model_name)

# Подготовка входного текста
input_text = "Напиши очень краткую рецензию о книге гарри поттер."

# Токенизация и генерация текста
input_ids = tokenizer.encode(input_text, return_tensors="pt")
output = model.generate(
  input_ids,
  max_length=1512,
  temperature=0.3,
  num_return_sequences=1,
  no_repeat_ngram_size=2,
  top_k=50,
  top_p=0.95,
  )

# Декодирование и вывод результата
generated_text = tokenizer.decode(output[0], skip_special_tokens=True)
print(generated_text)
