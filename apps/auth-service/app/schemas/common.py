from pydantic import BaseModel
from typing import Optional, Generic, TypeVar, Any

T = TypeVar("T")

class MessageResponse(BaseModel):
    code: int
    message: str
    data: Optional[Any] = None
