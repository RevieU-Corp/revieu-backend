from pydantic import BaseModel
from typing import Optional, TypeVar, Any

T = TypeVar("T")


class MessageResponse(BaseModel):
    code: int
    message: str
    data: Optional[Any] = None
