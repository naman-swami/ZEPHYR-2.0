"""
=============================================================================
  ZEPHYR 2.0: Kaggle Notebook for Qwen2.5-Coder-0.5B Multi-Task Fine-Tuning
=============================================================================

Follow these simple steps on Kaggle:
1. Go to https://www.kaggle.com/code and click "New Notebook".
2. In the right-hand panel:
   - Accelerator: Select "GPU T4 x 2" (or "GPU P100").
   - Internet: Toggle "Internet on" (Required to download base model).
3. Upload the dataset:
   - Click "Add Input" -> "Upload" -> Select "ai/data/zephyr_multitask_70k.jsonl".
4. Copy and run the code cells below.
"""

# =============================================================================
# CELL 1: Install Required Libraries (1 minute)
# =============================================================================
"""
!pip install -q -U torch transformers peft trl datasets bitsandbytes accelerate
"""

# =============================================================================
# CELL 2: Training Pipeline Setup (30 seconds)
# =============================================================================
"""
import os
import torch
from transformers import (
    AutoModelForCausalLM,
    AutoTokenizer,
    BitsAndBytesConfig,
    TrainingArguments
)
from peft import LoraConfig, TaskType, get_peft_model
from datasets import load_dataset
from trl import SFTTrainer, DataCollatorForCompletionOnlyLM

# Model Identifier
MODEL_ID = "Qwen/Qwen2.5-Coder-0.5B-Instruct"

# Find dataset path on Kaggle
dataset_candidates = [
    "/kaggle/input/zephyr-multitask-70k/zephyr_multitask_70k.jsonl",
    "/kaggle/input/zephyr_multitask_70k.jsonl",
    "./zephyr_multitask_70k.jsonl"
]

DATASET_PATH = None
for candidate in dataset_candidates:
    if os.path.exists(candidate):
        DATASET_PATH = candidate
        break

if not DATASET_PATH:
    # Look for any jsonl file in kaggle input
    for root, _, files in os.walk("/kaggle/input"):
        for f in files:
            if f.endswith(".jsonl"):
                DATASET_PATH = os.path.join(root, f)
                break

print(f"[OK] Using Dataset: {DATASET_PATH}")
"""

# =============================================================================
# CELL 3: Load Model with 4-bit NF4 Quantization
# =============================================================================
"""
bnb_config = BitsAndBytesConfig(
    load_in_4bit=True,
    bnb_4bit_quant_type="nf4",
    bnb_4bit_use_double_quant=True,
    bnb_4bit_compute_dtype=torch.bfloat16
)

print("Downloading and quantizing base model...")
tokenizer = AutoTokenizer.from_pretrained(MODEL_ID, trust_remote_code=True)
tokenizer.pad_token = tokenizer.eos_token

model = AutoModelForCausalLM.from_pretrained(
    MODEL_ID,
    quantization_config=bnb_config,
    torch_dtype=torch.bfloat16,
    device_map="auto",
    trust_remote_code=True
)
print("[OK] Base model loaded in 4-bit precision (~350 MB VRAM)!")
"""

# =============================================================================
# CELL 4: Configure LoRA and Training Parameters
# =============================================================================
"""
lora_config = LoraConfig(
    task_type=TaskType.CAUSAL_LM,
    r=32,
    lora_alpha=64,
    lora_dropout=0.05,
    bias="none",
    target_modules=[
        "q_proj", "k_proj", "v_proj", "o_proj",
        "gate_proj", "up_proj", "down_proj"
    ]
)

training_args = TrainingArguments(
    output_dir="/kaggle/working/zephyr_coder_checkpoints",
    per_device_train_batch_size=16,
    gradient_accumulation_steps=4,      # Effective batch size = 64
    learning_rate=2.0e-4,
    lr_scheduler_type="cosine",
    warmup_ratio=0.05,
    num_train_epochs=3,
    logging_steps=50,
    save_strategy="epoch",
    bf16=True,
    fp16=False,
    optim="paged_adamw_8bit",
    weight_decay=0.01,
    gradient_checkpointing=True,
    report_to="none"
)

# Completion-only loss: compute loss strictly on the model's generated diff or fix!
response_template = "<|im_start|>assistant\n"
collator = DataCollatorForCompletionOnlyLM(
    response_template=response_template,
    tokenizer=tokenizer
)

raw_dataset = load_dataset("json", data_files=DATASET_PATH, split="train")
print(f"[OK] Loaded {len(raw_dataset)} training examples!")

trainer = SFTTrainer(
    model=model,
    train_dataset=raw_dataset,
    dataset_text_field="prompt",
    max_seq_length=1024,
    data_collator=collator,
    args=training_args,
    peft_config=lora_config
)
"""

# =============================================================================
# CELL 5: Launch Training (Takes ~25-35 minutes on Kaggle T4)
# =============================================================================
"""
print("Starting fine-tuning...")
trainer.train()

OUTPUT_ADAPTER_DIR = "/kaggle/working/zephyr_coder_adapter"
trainer.model.save_pretrained(OUTPUT_ADAPTER_DIR)
tokenizer.save_pretrained(OUTPUT_ADAPTER_DIR)
print(f"[SUCCESS] LoRA adapter successfully saved to: {OUTPUT_ADAPTER_DIR}")
"""

# =============================================================================
# CELL 6: Zip the Adapter for One-Click Download
# =============================================================================
"""
import shutil
shutil.make_archive("/kaggle/working/zephyr_coder_adapter", "zip", OUTPUT_ADAPTER_DIR)
print("[READY] Download 'zephyr_coder_adapter.zip' from the Kaggle Output tab!")
"""
